package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/FileConversionApi/utils"
)

func main() {

	config, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config")
	}

	if config.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("Hello, world!"))
	}))

	httpServer := &http.Server{
		Addr:    config.HttpServerAddress,
		Handler: mux,
	}

	// Start the HTTPS server
	fmt.Printf("Starting server on %s\n", config.HttpServerAddress)
	certFile := filepath.Join(config.CertPath, config.CertFile)
	certKey := filepath.Join(config.CertPath, config.CertKey)

	err = httpServer.ListenAndServeTLS(certFile, certKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}
