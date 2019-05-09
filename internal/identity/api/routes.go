package api

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/decentraland/world/internal/commons/token"
	"github.com/decentraland/world/internal/commons/utils"
	"github.com/decentraland/world/internal/gindcl"
	"github.com/decentraland/world/internal/identity/data"
	"github.com/decentraland/world/internal/identity/repository"
	"github.com/decentraland/world/internal/identity/validation"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gopkg.in/go-playground/validator.v9"
)

type Application struct {
	auth0            data.IAuth0Service
	redirectURL      string
	generator        *token.Generator // Signs jwt, maybe should update name?s
	pubkey           string
	serverUrl        string
	clientRepository repository.ClientRepository
	validator        *validator.Validate
}

type TokenRequest struct {
	UserToken string `json:"user_token"`
	PublicKey string `json:"pub_key"`
}

type AuthRequest struct {
	Domain    string `json:"domain" validate:"required"`
	LoginUrl  string `json:"login_callback" validate:"required"`
	LogoutUrl string `json:"logout_callback" validate:"required"`
}

type AuthResponse struct {
	LoginUrl  string `json:"login_url"`
	LogoutUrl string `json:"logout_url"`
}

func InitApi(auth0Service data.IAuth0Service, key *ecdsa.PrivateKey, router *gin.Engine, clientRepository repository.ClientRepository, serverUrl string, jwtDuration time.Duration) error {
	generator := token.New(key, "1.0", jwtDuration)

	publicKey, err := utils.PemEncodePublicKey(&key.PublicKey)
	if err != nil {
		return err
	}

	router.Use(gindcl.CorsMiddleware())

	app := &Application{auth0: auth0Service, generator: generator, pubkey: publicKey, clientRepository: clientRepository, serverUrl: serverUrl, validator: validator.New()}

	api := router.Group("/api")

	v1 := api.Group("/v1")

	v1.GET("/status", app.status)
	v1.GET("/public_key", app.publicKey)
	v1.POST("/auth", app.authData)
	v1.POST("/token", app.token)

	// Handle pre-flight checks one by one
	v1.OPTIONS("/public_key", gindcl.PrefligthChecksMiddleware("GET", "*"))
	v1.OPTIONS("/auth", gindcl.PrefligthChecksMiddleware("POST", "*"))
	v1.OPTIONS("/token", gindcl.PrefligthChecksMiddleware("POST", "*"))

	return nil
}

func (a *Application) status(c *gin.Context) {
	c.String(http.StatusOK, "available")
}

func (a *Application) publicKey(c *gin.Context) {
	c.String(http.StatusOK, a.pubkey)
}

func (a *Application) authData(c *gin.Context) {
	var req AuthRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := a.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}

	response, err := getAuthData(a.clientRepository, req, a.serverUrl)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (a *Application) token(c *gin.Context) {
	var params TokenRequest
	err := c.ShouldBindJSON(&params)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad parameters, need user_token and pub_key"})
		return
	}

	if err := a.validator.Struct(params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}

	if !validation.ValidateEphemeralKey(params.PublicKey) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "public key found, but looks invalid"})
		return
	}

	user, err := a.auth0.GetUserInfo(params.UserToken)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	token, err := a.generator.NewToken(user.UserId, params.PublicKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, "error signing token")
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": token})
}

func buildUrl(basePath string, relPath string, args ...interface{}) string {
	u, _ := url.Parse(basePath)
	u.Path = path.Join(u.Path, fmt.Sprintf(relPath, args...))
	urlResult, _ := url.PathUnescape(u.String())
	return urlResult
}

func getAuthData(clientRepository repository.ClientRepository, req AuthRequest, serverUrl string) (*AuthResponse, error) {
	client, err := clientRepository.GetByDomain(req.Domain)
	if err != nil {
		log.WithError(err).Error("Error retrieving client data by domain")
		return nil, err
	}

	if client.LogoutUrl != req.LogoutUrl || client.LoginUrl != req.LoginUrl {
		log.Error("Provided data do no match")
		return nil, errors.New("provided data do no match")
	}

	return &AuthResponse{
		LoginUrl:  buildUrl(serverUrl, "/login/%s", client.Id),
		LogoutUrl: buildUrl(serverUrl, "/logout/%s", client.Id),
	}, nil
}
