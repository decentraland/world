package api

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/decentraland/world/internal/commons/version"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/decentraland/world/internal/commons/utils"
	"github.com/decentraland/world/internal/identity/data"
	"github.com/decentraland/world/internal/identity/repository"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gopkg.in/go-playground/validator.v9"
)

type application struct {
	auth0            data.IAuth0Service
	redirectURL      string
	generator        *utils.TokenGenerator // Signs jwt, maybe should update name?s
	pubkey           string
	serverURL        string
	clientRepository repository.ClientRepository
	validator        *validator.Validate
}

type tokenRequest struct {
	UserToken string `json:"user_token" validate:"required"`
	PublicKey string `json:"pub_key" validate:"required"`
}

type authRequest struct {
	Domain    string `json:"domain" validate:"required"`
	LoginURL  string `json:"login_callback" validate:"required"`
	LogoutURL string `json:"logout_callback" validate:"required"`
}

type authResponse struct {
	LoginURL  string `json:"login_url"`
	LogoutURL string `json:"logout_url"`
}

// Config is the API config
type Config struct {
	Auth0Service     data.IAuth0Service
	Key              *ecdsa.PrivateKey
	ClientRepository repository.ClientRepository
	ServerURL        string
	JWTDuration      time.Duration
}

// InitAPI initializes api routes
func InitAPI(router *gin.Engine, config *Config) error {
	generator := utils.NewTokenGenerator(config.Key, "1.0", config.JWTDuration)

	publicKey, err := utils.PemEncodePublicKey(&config.Key.PublicKey)
	if err != nil {
		return err
	}

	router.Use(utils.CorsMiddleware())

	app := &application{
		auth0:            config.Auth0Service,
		generator:        generator,
		pubkey:           publicKey,
		clientRepository: config.ClientRepository,
		serverURL:        config.ServerURL,
		validator:        validator.New(),
	}

	api := router.Group("/api")

	v1 := api.Group("/v1")

	version.RegisterVersionEndpoint(v1)

	v1.GET("/status", app.status)
	v1.GET("/public_key", app.publicKey)
	v1.POST("/auth", app.authData)
	v1.POST("/token", app.token)

	// Handle pre-flight checks one by one
	v1.OPTIONS("/public_key", utils.PrefligthChecksMiddleware("GET", utils.AllHeaders))
	v1.OPTIONS("/auth", utils.PrefligthChecksMiddleware("POST", utils.AllHeaders))
	v1.OPTIONS("/token", utils.PrefligthChecksMiddleware("POST", utils.AllHeaders))

	return nil
}

func (a *application) status(c *gin.Context) {
	c.String(http.StatusOK, "available")
}

func (a *application) publicKey(c *gin.Context) {
	c.String(http.StatusOK, a.pubkey)
}

func (a *application) authData(c *gin.Context) {
	var req authRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := a.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}

	client, err := a.clientRepository.GetByDomain(req.Domain)
	if err != nil {
		log.WithError(err).Error("Error retrieving client data by domain")
		c.JSON(http.StatusUnauthorized, gin.H{"error": errors.New("Error retrieving data")})
		return
	}

	if client.LogoutURL != req.LogoutURL || client.LoginURL != req.LoginURL {
		log.Error("Provided data do no match")
		c.JSON(http.StatusUnauthorized, gin.H{"error": errors.New("Provided data do no match")})
		return
	}

	response := &authResponse{
		LoginURL:  buildURL(a.serverURL, "/login/%s", client.ID),
		LogoutURL: buildURL(a.serverURL, "/logout/%s", client.ID),
	}

	c.JSON(http.StatusOK, response)
}

func (a *application) token(c *gin.Context) {
	var params tokenRequest
	err := c.ShouldBindJSON(&params)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad parameters, need user_token and pub_key"})
		return
	}

	if err := a.validator.Struct(params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}

	user, err := a.auth0.GetUserInfo(params.UserToken)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	token, err := a.generator.NewToken(user.UserID, params.PublicKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, "error signing token")
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": token})
}

func buildURL(basePath string, relPath string, args ...interface{}) string {
	u, _ := url.Parse(basePath)
	u.Path = path.Join(u.Path, fmt.Sprintf(relPath, args...))
	urlResult, _ := url.PathUnescape(u.String())
	return urlResult
}
