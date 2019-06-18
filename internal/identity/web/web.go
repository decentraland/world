package web

import (
	"net/http"
	"net/url"
	"path"

	"github.com/decentraland/world/internal/identity/repository"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func SiteContent(router *gin.Engine, clientRepo repository.ClientRepository, serverURL string, auth0Domain string) {
	router.Use(static.Serve("/public", static.LocalFile("internal/identity/web/static/", true)))
	router.LoadHTMLGlob("internal/identity/web/templates/*")

	app := &application{
		clientRepo:  clientRepo,
		serverURL:   serverURL,
		authODomain: auth0Domain,
	}

	router.GET("/login/:clientId", app.login)
	router.GET("/login_callback", app.loginCallback)

	router.GET("/logout/:clientId", app.logout)
	router.GET("/logout_callback", app.logoutCallback)

}

type application struct {
	clientRepo  repository.ClientRepository
	serverURL   string
	authODomain string
}

func (app *application) buildCallbackURL(relURL string, clientID string) (string, error) {
	u, err := url.Parse(app.serverURL)
	if err != nil {
		log.WithError(err).Errorf("Failed to parse serverURL %s", app.serverURL)
		return "", err
	}

	u.Path = path.Join(u.Path, "/login_callback")
	v := url.Values{}
	v.Set("clientId", clientID)
	u.RawQuery = v.Encode()

	return u.String(), nil
}

func (app *application) login(c *gin.Context) {
	clientID := c.Param("clientId")

	if client, err := app.clientRepo.GetByID(clientID); err != nil {
		log.WithError(err).Errorf("Failed to retrieve client by id[%s]", clientID)
		c.HTML(http.StatusBadRequest, "errorPage.html", gin.H{
			"message": err.Error()})
	} else {
		callbackURL, err := app.buildCallbackURL("/login_callback", clientID)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "errorPage.html", gin.H{
				"message": "Internal Server Error"})
			return
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", client.Domain)
		c.HTML(http.StatusOK, "login.html", gin.H{
			"callbackUrl": callbackURL,
			"domain":      app.authODomain,
			"externalId":  client.ExternalID,
		})
	}
}

func (app *application) loginCallback(c *gin.Context) {
	params := c.Request.URL.Query()

	data, ok := params["clientId"]
	if !ok {
		c.HTML(http.StatusBadRequest, "errorPage.html", gin.H{
			"message": "Invalid url"})
		return
	}

	clientID := data[0]

	if client, err := app.clientRepo.GetByID(clientID); err != nil {
		log.WithError(err).Errorf("Failed to retrieve client by id[%s]", clientID)
		c.HTML(http.StatusBadRequest, "errorPage.html", gin.H{
			"message": err.Error()})
	} else {
		callbackURL, err := app.buildCallbackURL("", clientID)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "errorPage.html", gin.H{
				"message": "Internal Server Error"})
			return
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", client.Domain)
		c.HTML(http.StatusOK, "loginCallback.html", gin.H{
			"callbackUrl": callbackURL,
			"authDomain":  app.authODomain,
			"externalId":  client.ExternalID,
			"redirectUrl": client.GetFullLoginURL(),
			"appDomain":   client.Domain,
		})
	}

}

func (app *application) logout(c *gin.Context) {
	clientID := c.Param("clientId")

	if client, err := app.clientRepo.GetByID(clientID); err != nil {
		log.WithError(err).Errorf("Failed to retrieve client by id[%s]", clientID)
		c.HTML(http.StatusBadRequest, "errorPage.html", gin.H{
			"message": err.Error()})
	} else {
		callbackURL, err := app.buildCallbackURL("/logout_callback", clientID)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "errorPage.html", gin.H{
				"message": "Internal Server Error"})
			return
		}
		c.Writer.Header().Set("Access-Control-Allow-Origin", client.Domain)
		c.HTML(http.StatusOK, "logout.html", gin.H{
			"callbackUrl": callbackURL,
			"domain":      app.authODomain,
			"externalId":  client.ExternalID,
		})
	}
}

func (app *application) logoutCallback(c *gin.Context) {
	params := c.Request.URL.Query()

	data, ok := params["clientId"]
	if !ok {
		c.HTML(http.StatusBadRequest, "errorPage.html", gin.H{
			"message": "Invalid url"})
		return
	}

	clientID := data[0]

	if client, err := app.clientRepo.GetByID(clientID); err != nil {
		log.WithError(err).Errorf("Failed to retrieve client by id[%s]", clientID)
		c.HTML(http.StatusBadRequest, "errorPage.html", gin.H{
			"message": err.Error()})
	} else {
		c.Writer.Header().Set("Access-Control-Allow-Origin", client.Domain)
		c.HTML(http.StatusOK, "logoutCallback.html", gin.H{
			"redirectUrl": client.GetFullLogoutURL(),
		})
	}
}
