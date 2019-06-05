package auth

import (
	"crypto/ecdsa"
	"fmt"
	"net/http"

	"github.com/decentraland/world/internal/commons/utils"

	auth2 "github.com/decentraland/auth-go/pkg/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MiddlewareConfiguration struct {
	AuthServerURL string
	RequestTTL    int64
}

func NewAuthMiddleware(c *MiddlewareConfiguration, publicURL string) (func(ctx *gin.Context), error) {
	pubKey, err := utils.ReadRemotePublicKey(c.AuthServerURL)
	if err != nil {
		return nil, fmt.Errorf("cannot read public key from '%s': %v", c.AuthServerURL, err)
	}

	return NewThirdPartyAuthMiddleware(pubKey, c.RequestTTL, publicURL)
}

func NewThirdPartyAuthMiddleware(pubKey *ecdsa.PublicKey, reqTTL int64, publicURL string) (func(ctx *gin.Context), error) {
	authHandler, err := auth2.NewThirdPartyAuthProvider(&auth2.ThirdPartyProviderConfig{
		RequestLifeSpan: reqTTL,
		TrustedKey:      pubKey,
	})
	if err != nil {
		return nil, err
	}

	return func(ctx *gin.Context) {
		if ctx.Request.Method == http.MethodOptions {
			ctx.Next()
		} else {
			//req, err := auth2.MakeFromHttpRequest(ctx.Request, publicURL)
			//if err != nil {
			//	ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unable to authenticate request"})
			//	return
			//}
			//ok, err := authHandler.ApproveRequest(req)
			//if err != nil {
			//	switch err := err.(type) {
			//	case auth2.AuthenticationError:
			//		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			//	case auth2.AuthorizationError:
			//		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": err.Error()})
			//	default:
			//		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			//	}
			//	return
			//}
			//if !ok {
			//	ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unable to authenticate request"})
			//	return
			//}
			ctx.Next()
		}
	}, nil
}

func IdExtractorMiddleware(ctx *gin.Context) {
	token, err := auth2.ExtractAuthTokenPayload(ctx.Request.Header.Get(auth2.HeaderAccessToken))
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
