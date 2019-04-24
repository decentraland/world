package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/decentraland/world/internal/auth"
	"os"

	"github.com/decentraland/world/internal/profile"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/toorop/gin-logrus"
)

func main() {
	host := flag.String("host", "localhost", "")
	port := flag.Int("port", 8081, "")
	connStr := flag.String("connStr", "", "psql connection string")
	schemaDir := flag.String("schemaDir", "pkg/profile", "Path to the directory containing json schema files")
	flag.Parse()

	log := logrus.New()
	router := gin.Default()
	router.Use(ginlogrus.Logger(log), gin.Recovery())

	if *connStr == "" {
		log.Fatal("missing --connStr")
	}

	db, err := sql.Open("postgres", *connStr)
	if err != nil {
		log.Fatal(err)
	}

	setupAuthentication(router)

	config := profile.Config{
		Services:  profile.Services{Log: log, Db: db},
		SchemaDir: *schemaDir,
	}
	err = profile.Register(&config, router)
	if err != nil {
		log.WithError(err).Fatal("unable to start profile service")
	}

	addr := fmt.Sprintf("%s:%d", *host, *port)
	router.Run(addr)
}

func setupAuthentication(r *gin.Engine) {
	authPubKey := os.Getenv("AUTH_KEY")
	authConfig := &auth.Configuration{Mode: auth.AuthThirdParty, AuthKey:authPubKey, RequestTTL: 60}

	r.Use(auth.NewAuthMiddleware(authConfig))
	r.Use(auth.IdExtractorMiddleware)
}