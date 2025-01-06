package main

import (
	"context"
	"os"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/FileConversionApi/api"
	db "github.com/FileConversionApi/db/sqlc"
	"github.com/FileConversionApi/utils"
	"github.com/FileConversionApi/workers"
)

func main() {
	config, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config")
	}

	if config.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	serverCtx, srvCancel := context.WithCancel(context.Background()) // Create context for the server
	defer srvCancel()

	// Database connection pool for server
	srvConnPool, err := pgxpool.New(serverCtx, config.ConnectionString)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}
	defer srvConnPool.Close()

	srvStore := db.NewStore(srvConnPool)

	// Context and cancel function for processor
	processorCtx, processorCancel := context.WithCancel(context.Background()) // Create context for the processor
	defer processorCancel()

	// Database connection pool for processor
	processorConnPool, err := pgxpool.New(serverCtx, config.ConnectionString)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}
	defer processorConnPool.Close()

	processorStore := db.NewStore(processorConnPool)

	// Signal handler setup to catch SIGINT, SIGTERM
	signalChan := api.SetupSignalHandler(srvCancel, processorCancel)

	// Database migration
	api.RunDBMigration(config.MigrationLocation, config.ConnectionString)

	storage := utils.LocalStorage{}
	errChan := make(chan error, 2)

	converter := utils.PdfConverter{}
	processor := workers.Builder().
		WithConverter(converter).
		WithCtx(processorCtx).
		WithStorage(storage).
		WithStore(processorStore).
		Build()

	gen := utils.NewJwtGenerator(config.SigningKey)
	server := api.
		Builder().
		WithCtx(serverCtx).
		WithStore(srvStore).
		WithConfig(config).
		WithStorage(storage).
		WithTokenGen(gen).
		Build()

	// Start background processor and server in goroutines
	go func() { errChan <- processor.Work() }()
	go func() { errChan <- server.Start() }()

	// Wait for shutdown signal or errors
	api.WaitForShutdown(errChan, signalChan, srvCancel, processorCancel)
	log.Info().Msg("Application shut down gracefully.")
}
