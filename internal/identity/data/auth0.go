package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const (
	requestTimeout = 2 * time.Second
)

type Auth0Config struct {
	Domain string
}

type User struct {
	Email  string
	UserID string
}

type IAuth0Service interface {
	GetUserInfo(accessToken string) (User, error)
}

type Auth0Service struct {
	baseURL        *url.URL
	getUserInfoURL string
}

func MakeAuth0Service(config Auth0Config) (IAuth0Service, error) {
	u, err := url.Parse(fmt.Sprintf("https://%s", config.Domain))
	if err != nil {
		return nil, err
	}

	getUserInfoRel, err := url.Parse("/userinfo")
	if err != nil {
		return nil, err
	}

	s := &Auth0Service{
		baseURL:        u,
		getUserInfoURL: u.ResolveReference(getUserInfoRel).String(),
	}

	return s, nil
}

type getUserInfoResponse struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
}

func (s *Auth0Service) GetUserInfo(accessToken string) (User, error) {
	user := User{}

	req, _ := http.NewRequest("GET", s.getUserInfoURL, nil)
	req.Header.Add("content-type", "application/json")
	req.Header.Add("authorization", fmt.Sprintf("Bearer %s", accessToken))
	client := http.Client{Timeout: requestTimeout}
	res, err := client.Do(req)
	if err != nil {
		return user, err
	}

	defer res.Body.Close()
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		getUserInfoResponse := &getUserInfoResponse{}
		if err := json.NewDecoder(res.Body).Decode(getUserInfoResponse); err != nil {
			return user, err
		}

		user.Email = getUserInfoResponse.Email
		user.UserID = getUserInfoResponse.Sub
		return user, nil
	}

	return user, handleErrorResponse(res)
}

func handleErrorResponse(response *http.Response) error {
	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	errorMsg := string(bodyBytes)

	msg := fmt.Sprintf("%d(%s) - %s", response.StatusCode, response.Status, errorMsg)

	switch response.StatusCode {
	case http.StatusUnauthorized:
		return UnauthorizedError{msg}
	case http.StatusTooManyRequests:
		return RateLimitError{msg}
	case http.StatusForbidden:
		return ForbiddenError{msg}
	case http.StatusBadRequest:
		return BadRequestError{msg}
	case http.StatusInternalServerError:
		return InternalError{msg}
	case http.StatusServiceUnavailable:
		return ServiceUnavailableError{msg}
	}
	return errors.New(msg)
}

type UnauthorizedError struct {
	Cause string
}

func (e UnauthorizedError) Error() string {
	return e.Cause
}

type ForbiddenError struct {
	Cause string
}

func (e ForbiddenError) Error() string {
	return e.Cause
}

type RateLimitError struct {
	Cause string
}

func (e RateLimitError) Error() string {
	return e.Cause
}

type ServiceUnavailableError struct {
	Cause string
}

func (e ServiceUnavailableError) Error() string {
	return e.Cause
}

type InternalError struct {
	Cause string
}

func (e InternalError) Error() string {
	return e.Cause
}

type BadRequestError struct {
	Cause string
}

func (e BadRequestError) Error() string {
	return e.Cause
}