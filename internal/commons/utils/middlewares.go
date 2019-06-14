package utils

import (
	"github.com/decentraland/auth-go/pkg/auth"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Because FireFox does not handle '*' in the Access-Control-Allow-Headers header
const BasicHeaders = "Accept,Accept-Encoding, Access-Control-Allow-Credentials, Access-Control-Allow-Headers, Access-Control-Allow-Methods, Access-Control-Allow-Origin, " +
	"Access-Control-Max-Age, Age, Allow,Authentication-Info, Authorization,CONNECT, Cache-Control, Connection, Content-Base, Content-Length, Content-Location, " +
	"Content-MD5, Content-Type, Content-Version, Cookie, DELETE,Destination, Expires, From, GET, HEAD, Host, Keep-Alive, Location, MIME-Version, OPTION, OPTIONS, " +
	"Optional, Origin, POST, PUT, Protocol, Proxy-Authenticate, Proxy-Authentication-Info, Proxy-Authorization, Proxy-Features, Public, Referer, Refresh, Resolver-Location, " +
	"Sec-Websocket-Extensions, Sec-Websocket-Key, Sec-Websocket-Origin, Sec-Websocket-Protocol, Sec-Websocket-Version, Security-Scheme, Server, Set-Cookie, et-Cookie2, SetProfile, " +
	"Status, Timeout, Title, URI, User-Agent, Version, WWW-Authenticate, X-Content-Duration, X-Content-Security-Policy, X-Content-Type-Options, X-CustomHeader, X-DNSPrefetch-Control, " +
	"X-Forwarded-For, X-Forwarded-Port, X-Forwarded-Proto, X-Frame-Options, X-Modified, X-OTHER, X-PING, X-Requested-With"

const DclHeaders = auth.HeaderIdentity + ", " +
	auth.HeaderTimestamp + ", " +
	auth.HeaderAccessToken + ", " +
	auth.HeaderSignature + ", " +
	auth.HeaderAuthType + ", " +
	auth.HeaderCert + ", " +
	auth.HeaderCertSignature

const AllHeaders = BasicHeaders + ", " + DclHeaders

func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Next()
	}
}

func PrefligthChecksMiddleware(allowedMethods string, allowedHeaders string) func(c *gin.Context) {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Methods", allowedMethods)
		c.Writer.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
		c.AbortWithStatus(http.StatusOK)
	}
}
