package cli

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/decentraland/auth-go/pkg/ephemeral"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

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
	IdentityURL string
	PubKey      string
}

func (a *Auth) GetAccessToken(userToken string) (string, error) {
	c := http.Client{
		Timeout: time.Second * 10,
	}

	jsonBuff, err := json.Marshal(map[string]string{
		"user_token": userToken,
		"pub_key":    a.PubKey,
	})

	postTokenURL, err := url.Parse(a.IdentityURL)
	if err != nil {
		return "", err
	}

	postTokenURL.Path = path.Join(postTokenURL.Path, "/token")

	resp, err := c.Post(postTokenURL.String(), "application/json", bytes.NewReader(jsonBuff))
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	respBuff, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		response := make(map[string]string)
		json.Unmarshal(respBuff, &response)
		err := fmt.Errorf("http error %s %s %d, %s", postTokenURL, resp.Status, resp.StatusCode,
			response["error"])
		return "", err
	}

	response := make(map[string]interface{})
	if err := json.Unmarshal(respBuff, &response); err != nil {
		return "", err
	}

	return response["service_token"].(string), nil
}

func ExecuteAuthFlow(auth0 *Auth0, auth *Auth) (string, error) {
	userToken, err := auth0.GetUserToken()
	if err != nil {
		fmt.Println("error getting auth0 token", err)
		return "", nil
	}

	fmt.Println("user token", userToken)

	accessToken, err := auth.GetAccessToken(userToken)
	if err != nil {
		fmt.Println("error getting access token", err)
		return "", nil
	}

	fmt.Println("access token", accessToken)

	return accessToken, nil
}

func ReadEphemeralKeyFromFile(path string) (*ephemeral.EphemeralKey, error) {
	key, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot load private key from file %s: %v", path, err)
	}

	kbs, err := hex.DecodeString(string(key))

	privateKey, err := crypto.ToECDSA(kbs)
	if err != nil {
		return nil, err
	}

	config := ephemeral.EphemeralKeyConfig{
		PrivateKey: privateKey,
	}
	return ephemeral.NewEphemeralKey(&config)
}

func EncodePublicKey(key *ephemeral.EphemeralKey) string {
	return hexutil.Encode(crypto.CompressPubkey(key.PublicKey()))
}
