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

type authAPIErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
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
		json.NewDecoder(res.Body).Decode(getUserInfoResponse)

		user.Email = getUserInfoResponse.Email
		user.UserID = getUserInfoResponse.Sub
		return user, nil
	}

	errorResponse := &authAPIErrorResponse{}
	json.NewDecoder(res.Body).Decode(errorResponse)

	msg := fmt.Sprintf("%d %s - %s", res.StatusCode, errorResponse.Error, errorResponse.ErrorDescription)
	return user, errors.New(msg)
}
