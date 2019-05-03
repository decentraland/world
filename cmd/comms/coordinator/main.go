package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/decentraland/webrtc-broker/pkg/authentication"
	"github.com/decentraland/webrtc-broker/pkg/coordinator"
	"github.com/decentraland/world/internal/commons/logging"
)

func main() {
	host := flag.String("host", "localhost", "")
	port := flag.Int("port", 9090, "")
	version := flag.String("version", "UNKNOWN", "")
	logLevel := flag.String("logLevel", "debug", "")
	noopAuthEnabled := flag.Bool("noopAuthEnabled", false, "")
	flag.Parse()

	log := logging.New()
	if err := logging.SetLevel(log, *logLevel); err != nil {
		log.Error("error setting log level")
		return
	}
	defer logging.LogPanic()

	auth := authentication.Make()
	if *noopAuthEnabled {
		auth.AddOrUpdateAuthenticator("noop", &authentication.NoopAuthenticator{})
	}

	config := coordinator.Config{
		ServerSelector: coordinator.MakeRandomServerSelector(),
		Auth:           auth,
		Log:            log,
	}
	state := coordinator.MakeState(&config)

	go coordinator.Process(state)

	mux := http.NewServeMux()
	coordinator.Register(state, mux)

	addr := fmt.Sprintf("%s:%d", *host, *port)
	log.Info("starting coordinator ", addr, "- version:", *version)
	log.Fatal(http.ListenAndServe(addr, mux))
}
