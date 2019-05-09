package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const expectedDomain = "http://google.com"
const expectedId = "1"

func TestNewClientRepository(t *testing.T) {
	_, err := NewClientRepository("test/resources/NOT_THE_FILE.json")
	assert.NotNil(t, err)

	repo, err := NewClientRepository("../../identity_test/resources/clients-test.json")
	assert.Nil(t, err)
	assert.NotNil(t, repo)
}

func TestClientRepoImpl_GetByDomain(t *testing.T) {
	repo, err := NewClientRepository("../../identity_test/resources/clients-test.json")

	_, err = repo.GetByDomain("INVALID DOMAIN")
	assert.NotNil(t, err)
	assert.Equal(t, "client not found", err.Error())

	client, err := repo.GetByDomain(expectedDomain)
	assert.Nil(t, err)
	checkExpectedClient(t, client)
}

func TestClientRepoImpl_GetById(t *testing.T) {
	repo, err := NewClientRepository("../../identity_test/resources/clients-test.json")

	_, err = repo.GetById("INVALID ID")
	assert.NotNil(t, err)
	assert.Equal(t, "client not found", err.Error())

	client, err := repo.GetById(expectedId)
	assert.Nil(t, err)
	checkExpectedClient(t, client)
}

func checkExpectedClient(t *testing.T, client *ClientData) {
	assert.Equal(t, expectedDomain, client.Domain)
	assert.Equal(t, "/login_callback", client.LoginURL)
	assert.Equal(t, "/logout_callback", client.LogoutURL)
	assert.Equal(t, "externalId", client.ExternalID)
	assert.Equal(t, expectedId, client.Id)
}
