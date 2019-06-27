package auth

import (
	"crypto/ecdsa"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/decentraland/world/internal/commons/utils"
	"github.com/sirupsen/logrus"

	auth2 "github.com/decentraland/auth-go/pkg/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MiddlewareConfiguration struct {
	IdentityURL string
	PublicURL   string
	RequestTTL  int64
	Log         *logrus.Logger
}

func NewAuthMiddleware(c *MiddlewareConfiguration) (func(ctx *gin.Context), error) {
	u, err := url.Parse(c.IdentityURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "/public_key")
	pubKey, err := utils.ReadRemotePublicKey(u.String())
	if err != nil {
		return nil, fmt.Errorf("cannot read public key from '%s': %v", u.String(), err)
	}

	return NewThirdPartyAuthMiddleware(pubKey, c)
}

func NewThirdPartyAuthMiddleware(pubKey *ecdsa.PublicKey, c *MiddlewareConfiguration) (func(ctx *gin.Context), error) {
	authHandler, err := auth2.NewThirdPartyAuthProvider(&auth2.ThirdPartyProviderConfig{
		RequestLifeSpan: c.RequestTTL,
		TrustedKey:      pubKey,
	})
	if err != nil {
		return nil, err
	}

	return func(ctx *gin.Context) {
		if ctx.Request.Method == http.MethodOptions {
			ctx.Next()
		} else {
			req, err := auth2.MakeFromHttpRequest(ctx.Request, c.PublicURL)
			if err != nil {
				c.Log.WithError(err).Error("unable to authenticate request")
				ctx.Error(err)
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unable to authenticate request"})
				return
			}
			ok, err := authHandler.ApproveRequest(req)
			if err != nil {
				c.Log.WithError(err).Debug("not authorized")
				switch err := err.(type) {
				case auth2.AuthenticationError:
					ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				case auth2.AuthorizationError:
					ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": err.Error()})
				default:
					ctx.Error(err)
					c.Log.WithError(err).Error("internal error")
					ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal error, retry again later"})
				}
				return
			}
			if !ok {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unable to authenticate request"})
				return
			}
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
