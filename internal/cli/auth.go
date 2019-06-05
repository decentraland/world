package cli

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
)

func ReadKey(path string) (*ecdsa.PrivateKey, error) {
	key, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	kbs, err := hex.DecodeString(string(key))
	return crypto.ToECDSA(kbs)
}

type Auth0 struct {
	Domain       string
	Email        string
	Password     string
	ClientID     string
	ClientSecret string
	Audience     string
}

func (a *Auth0) GetUserToken() (string, error) {
	c := http.Client{
		Timeout: time.Second * 10,
	}

	postTokenURL := fmt.Sprintf("https://%s/oauth/token", a.Domain)
	payload := url.Values{}
	payload.Set("grant_type", "password")
	payload.Set("client_id", a.ClientID)
	payload.Set("client_secret", a.ClientSecret)
	payload.Set("audience", a.Audience)
	payload.Set("username", a.Email)
	payload.Set("password", a.Password)
	payload.Set("scope", "openid")
	resp, err := c.PostForm(postTokenURL, payload)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	respBuff, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	response := make(map[string]interface{})
	if err := json.Unmarshal(respBuff, &response); err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("http error %s %s %d, %s", postTokenURL, resp.Status, resp.StatusCode,
			response["error_description"])
		return "", err
	}

	return response["access_token"].(string), nil
}

type Auth struct {
	AuthBaseURL string
	UserToken   string
	PubKey      string
}

func (a *Auth) GetAccessToken() (string, error) {
	c := http.Client{
		Timeout: time.Second * 10,
	}

	jsonBuff, err := json.Marshal(map[string]string{
		"user_token": a.UserToken,
		"pub_key":    a.PubKey,
	})

	postTokenURL := fmt.Sprintf("%s/api/v1/token", a.AuthBaseURL)
	resp, err := c.Post(postTokenURL, "application/json", bytes.NewReader(jsonBuff))
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	respBuff, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	response := make(map[string]interface{})
	if err := json.Unmarshal(respBuff, &response); err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("http error %s %s %d, %s", postTokenURL, resp.Status, resp.StatusCode,
			response["error"])
		return "", err
	}

	return response["access_token"].(string), nil
}
