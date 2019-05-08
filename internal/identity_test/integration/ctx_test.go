package integration

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"github.com/decentraland/world/internal/identity/api"
	"github.com/decentraland/world/internal/identity/data"
	"github.com/decentraland/world/internal/identity/mocks"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type TestContext struct {
	User       TestUser
	Controller *gomock.Controller
	Router     *gin.Engine
	T          *testing.T
	Key        *ecdsa.PrivateKey
	Auth0      *mocks.MockIAuth0Service
}

type TestStep struct {
	Description string
	Line        string
	Step        func(ctx *TestContext) bool
}

type TestCase struct {
	Name  string
	Steps []TestStep
}

func tests(tcs ...TestCase) []TestCase {
	return tcs
}

func test(name string, fs ...TestStep) TestCase {
	return TestCase{Name: name, Steps: fs}
}

func s(description string, f func(ctx *TestContext) bool) TestStep {
	_, file, line, _ := runtime.Caller(1)
	return TestStep{
		Description: description,
		Step:        f,
		Line:        fmt.Sprintf("%s:%d", file, line),
	}
}

func RunTest(ctx *TestContext, test *TestCase) bool {
	passed := true
	failedAt := 0
	for i, step := range test.Steps {
		passed = step.Step(ctx)
		if !passed {
			failedAt = i
			break
		}
		i++
	}
	if passed {
		fmt.Println(test.Name + " OK")
	} else {
		fmt.Println(test.Name + " failed at " + test.Steps[failedAt].Description + "\n\t\tfrom: " + test.Steps[failedAt].Line)
	}
	return passed
}

func NewContext(t *testing.T) *TestContext {

	// disable logging
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard

	ctx := TestContext{}
	ctx.User = ValidUser()
	ctx.Controller = gomock.NewController(t)
	ctx.T = t
	ctx.Auth0 = mocks.NewMockIAuth0Service(ctx.Controller)
	ctx.Key, _ = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	ctx.Router = gin.New()
	if err := api.InitApi(ctx.Auth0, ctx.Key, ctx.Router, nil, "http://localhost:9091/", 60); err != nil {
		t.Fatal("Fail to initialize routes")
	}
	return &ctx
}

func SetupMocks(ctx *TestContext) {
	s := ctx.User.Token
	user := ctx.User.User()
	ctx.Auth0.EXPECT().GetUserInfo(s).Return(user, nil).AnyTimes()
}

func TimeTravel(ctx *TestContext) bool {
	ctx.Router = gin.New()
	api.InitApi(ctx.Auth0, ctx.Key, ctx.Router, nil, "\"http://localhost:9091/", time.Duration(-1))
	return true
}

func DefaultUser(ctx *TestContext) bool {
	ctx.User = ValidUser()
	return true
}

func InvalidUser(ctx *TestContext) bool {
	ctx.User.Token = "invalid_token"
	ctx.Auth0.EXPECT().GetUserInfo(ctx.User.Token).Return(data.User{}, errors.New("Unknown error"))
	return true
}

func UserWithKey(key string) func(ctx *TestContext) bool {
	return func(ctx *TestContext) bool {
		ctx.User = ValidUser()
		ctx.User.EphemeralKey = key
		return true
	}
}

func CallRefresh(ctx *TestContext) bool {
	w := httptest.NewRecorder()

	params := `{"user_token":"` + ctx.User.Token + `", "pub_key":"` + ctx.User.EphemeralKey + `"}`
	req, _ := http.NewRequest("POST", "/api/v1/token", bytes.NewReader([]byte(params)))

	ctx.Router.ServeHTTP(w, req)
	if w.Code != 200 {
		ctx.User.SessionJWT = ""
		return true
	}

	jwt := getJSONValue([]byte(w.Body.String()), "access_token")
	ok := assert.True(ctx.T, jwt != "", "missing token in 200 response")
	ctx.User.SessionJWT = jwt
	return ok
}

func InvalidRefresh(ctx *TestContext) bool {
	w := httptest.NewRecorder()

	params := `{"access_token":"` + ctx.User.Token + `", "ephemeral_key":"` + ctx.User.EphemeralKey + `"}`
	req, _ := http.NewRequest("POST", "/api/v1/token", bytes.NewReader([]byte(params)))

	ctx.Router.ServeHTTP(w, req)
	if w.Code == 200 {
		return false
	}

	jwt := getJSONValue([]byte(w.Body.String()), "token")
	ok := assert.True(ctx.T, jwt == "", "there shouldn't be any token")
	return ok
}

func AlterKey(ctx *TestContext) bool {
	ctx.Key, _ = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	return true
}

func ValidSessionJWT(ctx *TestContext) bool {
	token, err := jwt.Parse(ctx.User.SessionJWT, getKeyJWT(ctx.Key.X, ctx.Key.Y))
	if err != nil {
		return false
	}
	claims := token.Claims.(jwt.MapClaims)
	if len(claims) != 4 {
		return false
	}
	claimsKey, _ := claims["ephemeral_key"].(string)
	claimsVersion, _ := claims["version"].(string)
	claimsUserId, _ := claims["user_id"].(string)
	return claimsUserId == ctx.User.Id &&
		claimsKey == ctx.User.EphemeralKey &&
		claimsVersion == "1.0"

}

func InvalidSessionJWT(error uint32) func(*TestContext) bool {
	return func(ctx *TestContext) bool {
		_, err := jwt.Parse(ctx.User.SessionJWT, getKeyJWT(ctx.Key.X, ctx.Key.Y))
		verr, ok := err.(*jwt.ValidationError)
		return ok && (verr.Errors&error == error)
	}
}

func TestRefresh(t *testing.T) {

	fail := false
	cases := tests(
		test("Good refresh call",
			s("Use default user", DefaultUser),
			s("Call refresh", CallRefresh),
			s("Full JWT validation", ValidSessionJWT),
		),
		test("JWT bad key",
			s("Alter context key", AlterKey),
			s("Use default user", DefaultUser),
			s("Call refresh", CallRefresh),
			s("Check invalid signature", InvalidSessionJWT(jwt.ValidationErrorSignatureInvalid)),
		),
		test("Expired JWT",
			s("Default user", DefaultUser),
			s("Time travel", TimeTravel),
			s("Call refresh", CallRefresh),
			s("Check expired token", InvalidSessionJWT(jwt.ValidationErrorExpired)),
		),
		test("Unknown user",
			s("Unknown user", InvalidUser),
			s("Call refresh", CallRefresh),
			s("Nil token", InvalidSessionJWT(jwt.ValidationErrorMalformed)),
		),
		test("Ephemeral Key all lower case",
			s("Use default user", UserWithKey("aaaaaaaaaabbbbbbbbbbccccccccccddddddddddaaaaaaaaaabbbbbbbbbbcccccc")),
			s("Call refresh", CallRefresh),
			s("Full JWT validation", ValidSessionJWT),
		),
		test("Ephemeral Key with all upper case",
			s("Use default user", UserWithKey("AAAAAAAAAABBBBBBBBBBCCCCCCCCCCDDDDDDDDDDAAAAAAAAAABBBBBBBBBBCCCCCC")),
			s("Call refresh", CallRefresh),
			s("Full JWT validation", ValidSessionJWT),
		),
	)

	for _, c := range cases {
		ctx := NewContext(t)
		SetupMocks(ctx)
		fail = !RunTest(ctx, &c) || fail
	}

	assert.False(t, fail, "")
}
