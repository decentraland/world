package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/decentraland/auth-go/pkg/ephemeral"
)

type ProfileClient struct {
	ProfileURL   string
	EphemeralKey *ephemeral.EphemeralKey
}

func (pc *ProfileClient) StoreProfile(accessToken string, body io.Reader) error {
	u := fmt.Sprintf("%s/api/v1/profile", pc.ProfileURL)
	req, err := http.NewRequest("POST", u, body)
	if err != nil {
		return err
	}

	if err = pc.EphemeralKey.AddRequestHeaders(req, accessToken); err != nil {
		return err
	}

	c := http.Client{
		Timeout: time.Second * 10,
	}

	resp, err := c.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	respBuff, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		response := make(map[string]interface{})
		if err := json.Unmarshal(respBuff, &response); err != nil {
			return err
		}

		err := fmt.Errorf("http error %s %s %d, %v", u, resp.Status, resp.StatusCode, response["errors"])
		return err
	}

	return nil
}

func (pc *ProfileClient) RetrieveProfile(accessToken string) (map[string]interface{}, error) {
	u := fmt.Sprintf("%s/api/v1/profile", pc.ProfileURL)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	if err = pc.EphemeralKey.AddRequestHeaders(req, accessToken); err != nil {
		return nil, err
	}

	c := http.Client{
		Timeout: time.Second * 10,
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	respBuff, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	response := make(map[string]interface{})
	if err := json.Unmarshal(respBuff, &response); err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("http error %s %s %d", u, resp.Status, resp.StatusCode)
		return nil, err
	}

	return response, nil
}
