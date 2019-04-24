package auth

import (
	"fmt"
	auth2 "github.com/decentraland/auth-go/pkg/auth"
	"github.com/decentraland/auth-go/pkg/authentication"
	"github.com/decentraland/auth-go/pkg/authorization"
	"github.com/decentraland/auth-go/pkg/keys"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"strings"
)

type Configuration struct {
	Mode       string
	AuthKey    string
	RequestTTL int64
}

const (
	AuthOff        = "off"
	AuthThirdParty = "third_party"
)

func NewAuthMiddleware(c *Configuration) (func(ctx *gin.Context), error) {
	switch strings.ToLower(c.Mode) {
	case AuthOff:
		return createMiddleWare(c)
	case AuthThirdParty:
		return dummyMiddleWare(c)
	default:
		return nil, fmt.Errorf("undefined authentication mode: %s", c.Mode)
	}
}

func createMiddleWare(c *Configuration) (func(ctx *gin.Context), error) {
	k, err := keys.PemDecodePublicKey(c.AuthKey)
	if err != nil {
		return nil, err
	}

	authnStrategy := &authentication.ThirdPartyStrategy{RequestLifeSpan: c.RequestTTL, TrustedKey: k}
	authHandler := auth2.NewAuthProvider(authnStrategy, &authorization.AllowAllStrategy{})

	return func(ctx *gin.Context) {
		req, err := auth2.MakeFromHttpRequest(ctx.Request)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unable to authenticate request"})
			return
		}
		ok, err := authHandler.ApproveRequest(req)
		if err != nil {
			handleError(err, ctx)
			return
		}
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unable to authenticate request"})
			return
		}
		ctx.Next()
	}, nil
}

func dummyMiddleWare(c *Configuration) (func(ctx *gin.Context), error) {
	return nil, nil
}

func handleError(err error, ctx *gin.Context) {
	switch err := err.(type) {
	case auth2.AuthenticationError:
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case auth2.AuthorizationError:
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": err.Error()})
	default:
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
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
