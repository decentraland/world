package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	requestTimeout = 2 * time.Second
)

type Auth0Config struct {
	BaseUrl string `overwrite-env:"AUTH0_BASE_URL"`
	Domain  string `overwrite-env:"AUTH0_DOMAIN"`
}

type User struct {
	Email  string
	UserId string
}

type IAuth0Service interface {
	GetUserInfo(accessToken string) (User, error)
}

type Auth0Service struct {
	baseUrl        *url.URL
	getUserInfoUrl string
}

func MakeAuth0Service(config Auth0Config) (IAuth0Service, error) {
	u, err := url.Parse(config.BaseUrl)
	if err != nil {
		return nil, err
	}

	getUserInfoRel, err := url.Parse("/userinfo")
	if err != nil {
		return nil, err
	}

	s := &Auth0Service{
		baseUrl:        u,
		getUserInfoUrl: u.ResolveReference(getUserInfoRel).String(),
	}

	return s, nil
}

type AuthApiErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (s *Auth0Service) getWithAuth(url string, token string) (*http.Response, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("content-type", "application/json")
	req.Header.Add("authorization", fmt.Sprintf("Bearer %s", token))
	client := http.Client{Timeout: requestTimeout}
	return client.Do(req)
}

type GetUserInfoResponse struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
}

func (s *Auth0Service) GetUserInfo(accessToken string) (User, error) {
	user := User{}

	res, err := s.getWithAuth(s.getUserInfoUrl, accessToken)
	if err != nil {
		return user, err
	}

	defer res.Body.Close()
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		getUserInfoResponse := &GetUserInfoResponse{}
		json.NewDecoder(res.Body).Decode(getUserInfoResponse)

		user.Email = getUserInfoResponse.Email
		user.UserId = getUserInfoResponse.Sub
		return user, nil
	} else {
		errorResponse := &AuthApiErrorResponse{}
		json.NewDecoder(res.Body).Decode(errorResponse)

		msg := fmt.Sprintf("%d %s - %s", res.StatusCode, errorResponse.Error, errorResponse.ErrorDescription)
		return user, errors.New(msg)
	}
}
