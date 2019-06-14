package api

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"github.com/decentraland/world/internal/commons/utils"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/decentraland/world/internal/identity/mocks"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestOptionsHeaderResponse(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	repo := mocks.NewMockClientRepository(mockController)
	auth0 := mocks.NewMockIAuth0Service(mockController)
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fail()
	}

	router := gin.Default()

	config := Config{
		Auth0Service:     auth0,
		Key:              key,
		ClientRepository: repo,
		ServerURL:        "auth.decentraland.zone",
		JWTDuration:      time.Second,
	}
	if err := InitAPI(router, &config); err != nil {
		t.Fail()
	}

	for _, tc := range optionsCalls {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			req, _ := http.NewRequest("OPTIONS", tc.url, nil)

			router.ServeHTTP(w, req)

			tc.resultAssertion(t, w)
		})
	}
}

type optionTestCase struct {
	name            string
	url             string
	resultAssertion func(t *testing.T, w *httptest.ResponseRecorder)
}

var optionsCalls = []optionTestCase{
	{
		name:            "Public Key",
		url:             "/api/v1/public_key",
		resultAssertion: assertOkResponse("GET", utils.AllHeaders, "*"),
	},
	{
		name:            "Auth",
		url:             "/api/v1/auth",
		resultAssertion: assertOkResponse("POST", utils.AllHeaders, "*"),
	},
	{
		name:            "Token",
		url:             "/api/v1/token",
		resultAssertion: assertOkResponse("POST", utils.AllHeaders, "*"),
	},
	{
		name:            "Invalid url",
		url:             "/api/v1/random",
		resultAssertion: assertNotFound(),
	},
}

func assertOkResponse(allowedMethods, allowedHeaders, allowedOrigin string) func(t *testing.T, w *httptest.ResponseRecorder) {
	return func(t *testing.T, w *httptest.ResponseRecorder) {
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, allowedHeaders, w.Header().Get("Access-Control-Allow-Headers"))
		assert.Equal(t, allowedMethods, w.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, allowedOrigin, w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func assertNotFound() func(t *testing.T, w *httptest.ResponseRecorder) {
	return func(t *testing.T, w *httptest.ResponseRecorder) {
		assert.Equal(t, http.StatusNotFound, w.Code)
	}
}
