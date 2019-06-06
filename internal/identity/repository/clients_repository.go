package repository

import (
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"path"

	log "github.com/sirupsen/logrus"
)

type ClientData struct {
	LoginURL   string `json:"login_url"`
	LogoutURL  string `json:"logout_url"`
	Domain     string `json:"domain"`
	ID         string `json:"id"`
	ExternalID string `json:"external_id"`
}

func (cd *ClientData) GetFullLoginURL() string {
	return cd.buildURL(cd.LoginURL)
}

func (cd *ClientData) GetFullLogoutURL() string {
	return cd.buildURL(cd.LogoutURL)
}

func (cd *ClientData) buildURL(relPath string) string {
	if len(relPath) == 0 {
		return cd.Domain
	}
	u, _ := url.Parse(cd.Domain)
	u.Path = path.Join(u.Path, relPath)
	urlResult, _ := url.PathUnescape(u.String())
	return urlResult
}

type ClientRepository interface {
	GetByID(clientID string) (*ClientData, error)
	GetByDomain(domain string) (*ClientData, error)
}

type clientRepoImpl struct {
	IDIndex     map[string]ClientData
	domainIndex map[string]ClientData
}

func (c *clientRepoImpl) GetByID(clientID string) (*ClientData, error) {
	return doQuery(clientID, c.IDIndex)
}

func (c *clientRepoImpl) GetByDomain(domain string) (*ClientData, error) {
	return doQuery(domain, c.domainIndex)
}

func doQuery(key string, index map[string]ClientData) (*ClientData, error) {
	client, ok := index[key]

	if !ok {
		return nil, errors.New("client not found")
	}

	return &client, nil
}

func NewClientRepository(dataPath string) (ClientRepository, error) {
	data, err := readClientData(dataPath)
	if err != nil {
		log.WithError(err).Errorf("Failed to read client data from file: %s", dataPath)
		return nil, err
	}
	ids := make(map[string]ClientData)
	domains := make(map[string]ClientData)

	for _, c := range data {
		log.Debugf("Loading data for Domain: %s  Id: %s", c.Domain, c.ID)
		if _, ok := ids[c.ID]; !ok {
			ids[c.ID] = c
		} else {
			log.Warnf("Duplicate ClientID: %s", c.ID)
		}

		if _, ok := ids[c.Domain]; !ok {
			domains[c.Domain] = c
		} else {
			log.Warnf("Duplicate Domain: %s", c.ID)
		}
	}

	return &clientRepoImpl{IDIndex: ids, domainIndex: domains}, nil
}

func readClientData(dataPath string) ([]ClientData, error) {
	var data []ClientData
	c, err := os.Open(dataPath)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	if err = json.NewDecoder(c).Decode(&data); err != nil {
		return nil, err
	} else {
		return data, nil
	}
}
