package auth

import (
	"crypto/ecdsa"
	"fmt"
	"net/http"
	"strings"

	"github.com/decentraland/world/internal/commons/utils"

	auth2 "github.com/decentraland/auth-go/pkg/auth"
	"github.com/decentraland/auth-go/pkg/authentication"
	"github.com/decentraland/auth-go/pkg/authorization"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	AuthOff        = "off"
	AuthThirdParty = "third-party"
)

type Configuration struct {
	Mode          string `overwrite-flag:"authMode" flag-usage:"off, third-party" validate:"required"`
	AuthServerUrl string `overwrite-flag:"authUrl" flag-usage:"path to the file containing the auth-service public key"`
	RequestTTL    int64  `overwrite-flag:"authTtl" flag-usage:"request time to live"`
}

func NewAuthMiddleware(c *Configuration) (func(ctx *gin.Context), error) {
	switch strings.ToLower(c.Mode) {
	case AuthOff:
		return nil, nil
	case AuthThirdParty:
		k, err := utils.ReadRemotePublicKey(c.AuthServerUrl)
		if err != nil {
			return nil, fmt.Errorf("cannot read public key from '%s': %v", c.AuthServerUrl, err)
		}
		return NewThirdPartyAuthMiddleware(k, c.RequestTTL)
	default:
		return nil, fmt.Errorf("undefined authentication mode: %s", c.Mode)
	}
}

func NewThirdPartyAuthMiddleware(pubKey *ecdsa.PublicKey, reqTtl int64) (func(ctx *gin.Context), error) {
	authnStrategy := &authentication.ThirdPartyStrategy{RequestLifeSpan: reqTtl, TrustedKey: pubKey}
	authHandler := auth2.NewAuthProvider(authnStrategy, &authorization.AllowAllStrategy{})

	return func(ctx *gin.Context) {
		req, err := auth2.MakeFromHttpRequest(ctx.Request)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unable to authenticate request"})
			return
		}
		ok, err := authHandler.ApproveRequest(req)
		if err != nil {
			switch err := err.(type) {
			case auth2.AuthenticationError:
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			case auth2.AuthorizationError:
				ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": err.Error()})
			default:
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unable to authenticate request"})
			return
		}
		ctx.Next()
	}, nil
}

func IdExtractorMiddleware(ctx *gin.Context) {
	token, err := authentication.ExtractAuthTokenPayload(ctx.Request.Header.Get(auth2.HeaderAccessToken))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "unable to extract id from request"})
		return
	}
	ctx.Set("userId", token.UserId)
	ctx.Next()
}

// You can send the id you want to use in the 'x-dummy-id' header
// otherwhise a random id will be generated
func DummyIdExtractor(ctx *gin.Context) {
	id := ctx.Request.Header.Get("x-dummy-id")
	if len(id) == 0 {
		id = uuid.New().String()
	}
	ctx.Set("userId", id)
	ctx.Next()
}
