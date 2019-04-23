package main

import (
	"flag"
	"fmt"

	"database/sql"

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

	services := profile.Services{Log: log, Db: db}
	err = profile.Register(services, router)
	if err != nil {
		log.WithError(err).Fatal("unable to start profile service")
	}

	addr := fmt.Sprintf("%s:%d", *host, *port)
	router.Run(addr)
}
