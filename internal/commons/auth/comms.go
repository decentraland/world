package auth

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"path"

	auth2 "github.com/decentraland/auth-go/pkg/auth"
	brokerProtocol "github.com/decentraland/webrtc-broker/pkg/protocol"
	"github.com/decentraland/world/internal/commons/utils"
	protocol "github.com/decentraland/world/pkg/protocol"
	"github.com/golang/protobuf/proto"
)

// AuthenticatorConfig is the authenticator configuration
type AuthenticatorConfig struct {
	CoordinatorURL string
	Secret         string
	IdentityURL    string
	RequestTTL     int64
	Log            *logrus.Logger
}

// Authenticator is the DCL world authenticator, secret will be shared between servers and the
// client will use the normal world identity
type Authenticator struct {
	secret        string
	provider      auth2.AuthProvider
	authServerURL string
	connectURL    string
	log           *logrus.Logger
}

func joinURL(base string, rel string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, rel)
	return u.String(), nil
}

func MakeAuthenticator(config *AuthenticatorConfig) (*Authenticator, error) {
	pubKeyURL, err := joinURL(config.IdentityURL, "/public_key")
	if err != nil {
		return nil, err
	}
	pubKey, err := utils.ReadRemotePublicKey(pubKeyURL)
	if err != nil {
		return nil, fmt.Errorf("cannot read public key from '%s': %v", pubKeyURL, err)
	}
	authProvider, err := auth2.NewThirdPartyAuthProvider(&auth2.ThirdPartyProviderConfig{
		RequestLifeSpan: config.RequestTTL,
		TrustedKey:      pubKey,
	})
	if err != nil {
		return nil, err
	}

	connectURL, err := joinURL(config.CoordinatorURL, "/connect")
	if err != nil {
		return nil, err
	}

	a := &Authenticator{
		secret:     config.Secret,
		provider:   authProvider,
		connectURL: connectURL,
		log:        config.Log,
	}

	return a, nil
}

// AuthenticateFromMessage validates an auth message
func (a *Authenticator) AuthenticateFromMessage(role brokerProtocol.Role, body []byte) (bool, error) {
	if role == brokerProtocol.Role_COMMUNICATION_SERVER {
		return a.secret == string(body), nil
	} else if role == brokerProtocol.Role_CLIENT {
		authData := protocol.AuthData{}
		if err := proto.Unmarshal(body, &authData); err != nil {
			return false, err
		}

		credentials := make(map[string]string)
		credentials["x-signature"] = authData.Signature
		credentials["x-identity"] = authData.Identity
		credentials["x-timestamp"] = authData.Timestamp
		credentials["x-auth-type"] = "third-party"
		credentials["x-access-token"] = authData.AccessToken

		req := auth2.AuthRequest{Credentials: credentials, Content: []byte{}}
		ok, err := a.provider.ApproveRequest(&req)
		if err != nil {
			a.log.WithError(err).Error("failed to validate request")
		}
		return ok, err
	} else {
		return false, nil
	}
}

// AuthenticateFromURL validates an a coordinator request using the endpoint url
func (a *Authenticator) AuthenticateFromURL(role brokerProtocol.Role, r *http.Request) (bool, error) {
	qs := r.URL.Query()

	if role == brokerProtocol.Role_COMMUNICATION_SERVER {
		return a.secret == qs.Get("secret"), nil
	} else if role == brokerProtocol.Role_CLIENT {
		credentials := make(map[string]string)
		credentials["x-signature"] = qs.Get("signature")
		credentials["x-identity"] = qs.Get("identity")
		credentials["x-timestamp"] = qs.Get("timestamp")
		credentials["x-auth-type"] = "third-party"
		credentials["x-access-token"] = qs.Get("access-token")

		content := fmt.Sprintf("GET:%s", a.connectURL)
		req := auth2.AuthRequest{Credentials: credentials, Content: []byte(content)}
		ok, err :=  a.provider.ApproveRequest(&req)
		if err != nil {
			a.log.WithError(err).Error("failed to validate request")
		}
		return ok, err
	} else {
		return false, nil
	}
}

func (a *Authenticator) GenerateServerAuthMessage() (*brokerProtocol.AuthMessage, error) {
	m := &brokerProtocol.AuthMessage{
		Type: brokerProtocol.MessageType_AUTH,
		Role: brokerProtocol.Role_COMMUNICATION_SERVER,
		Body: []byte(a.secret),
	}
	return m, nil
}

func (a *Authenticator) GenerateServerConnectURL(coordinatorURL string) (string, error) {
	u := fmt.Sprintf("%s?secret=%s", coordinatorURL, a.secret)
	return u, nil
}
