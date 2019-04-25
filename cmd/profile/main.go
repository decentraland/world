package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	"github.com/decentraland/world/internal/auth"
	"github.com/decentraland/world/internal/profile"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	ginlogrus "github.com/toorop/gin-logrus"
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

	if err := setupAuthentication(router); err != nil {
		log.Fatal(err)
	}

	config := profile.Config{
		Services:  profile.Services{Log: log, Db: db},
		SchemaDir: *schemaDir,
	}
	err = profile.Register(&config, router)
	if err != nil {
		log.WithError(err).Fatal("unable to start profile service")
	}

	addr := fmt.Sprintf("%s:%d", *host, *port)

	if err := router.Run(addr); err != nil {
		log.WithError(err).Fatal("Fail to start server.")
	}
}

func setupAuthentication(r *gin.Engine) error {
	authPubKey := os.Getenv("AUTH_KEY")
	authConfig := &auth.Configuration{Mode: auth.AuthThirdParty, AuthKey: authPubKey, RequestTTL: 60}

	authMiddleware, err := auth.NewAuthMiddleware(authConfig)
	if err != nil {
		return err
	}
	r.Use(authMiddleware)
	r.Use(auth.IdExtractorMiddleware)
	return nil
}
