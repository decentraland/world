package web

import (
	"fmt"
	"net/http"

	"github.com/decentraland/world/internal/identity/repository"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func SiteContent(router *gin.Engine, clientRepo repository.ClientRepository, serverUrl string, auth0Domain string) {
	router.Use(static.Serve("/public", static.LocalFile("internal/identity/web/static/", true)))
	router.LoadHTMLGlob("internal/identity/web/templates/*")

	app := &application{
		clientRepo:  clientRepo,
		serverUrl:   serverUrl,
		authODomain: auth0Domain,
	}

	router.GET("/login/:clientId", app.login)
	router.GET("/login_callback", app.loginCallback)

	router.GET("/logout/:clientId", app.logout)
	router.GET("/logout_callback", app.logoutCallback)

}

type application struct {
	clientRepo  repository.ClientRepository
	serverUrl   string
	authODomain string
}

func (app *application) login(c *gin.Context) {

	clientId := c.Param("clientId")

	if client, err := app.clientRepo.GetById(clientId); err != nil {
		log.WithError(err).Errorf("Failed to retrieve client by id[%s]", clientId)
		c.HTML(http.StatusBadRequest, "errorPage.html", gin.H{
			"message": err.Error()})
	} else {
		callbackUrl := fmt.Sprintf("%slogin_callback?clientId=%s", app.serverUrl, clientId)
		c.Writer.Header().Set("Access-Control-Allow-Origin", client.Domain)
		c.HTML(http.StatusOK, "login.html", gin.H{
			"callbackUrl": callbackUrl,
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

	clientId := data[0]

	if client, err := app.clientRepo.GetById(clientId); err != nil {
		log.WithError(err).Errorf("Failed to retrieve client by id[%s]", clientId)
		c.HTML(http.StatusBadRequest, "errorPage.html", gin.H{
			"message": err.Error()})
	} else {
		c.Writer.Header().Set("Access-Control-Allow-Origin", client.Domain)
		callbackUrl := fmt.Sprintf("%s?clientId=%s", app.serverUrl, clientId)
		c.Writer.Header().Set("Access-Control-Allow-Origin", client.Domain)
		c.HTML(http.StatusOK, "loginCallback.html", gin.H{
			"callbackUrl": callbackUrl,
			"authDomain":  app.authODomain,
			"externalId":  client.ExternalID,
			"redirectUrl": client.GetFullLoginUrl(),
			"appDomain":   client.Domain,
		})
	}

}

func (app *application) logout(c *gin.Context) {

	clientId := c.Param("clientId")

	if client, err := app.clientRepo.GetById(clientId); err != nil {
		log.WithError(err).Errorf("Failed to retrieve client by id[%s]", clientId)
		c.HTML(http.StatusBadRequest, "errorPage.html", gin.H{
			"message": err.Error()})
	} else {
		callbackUrl := fmt.Sprintf("%slogout_callback?clientId=%s", app.serverUrl, clientId)
		c.Writer.Header().Set("Access-Control-Allow-Origin", client.Domain)
		c.HTML(http.StatusOK, "logout.html", gin.H{
			"callbackUrl": callbackUrl,
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

	clientId := data[0]

	if client, err := app.clientRepo.GetById(clientId); err != nil {
		log.WithError(err).Errorf("Failed to retrieve client by id[%s]", clientId)
		c.HTML(http.StatusBadRequest, "errorPage.html", gin.H{
			"message": err.Error()})
	} else {
		c.Writer.Header().Set("Access-Control-Allow-Origin", client.Domain)
		c.HTML(http.StatusOK, "logoutCallback.html", gin.H{
			"redirectUrl": client.GetFullLogoutUrl(),
		})
	}
}
