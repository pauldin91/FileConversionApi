package main

import (
	"context"
	"os"

	"github.com/FileConversionApi/api"
	"github.com/FileConversionApi/utils"
	"github.com/FileConversionApi/workers"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	srvChan := make(chan error)
	prcChan := make(chan error)

	config, err := utils.LoadConfig(".")

	storage := utils.LocalStorage{}
	converter := utils.PdfConverter{}
	gen := utils.NewJwtGenerator(config.SigningKey)

	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config")
	}

	if config.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := pgxpool.New(ctx, config.ConnectionString)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}
	defer pool.Close()

	// Database migration

	processor := workers.Builder().
		WithConverter(converter).
		WithStorage(storage).
		WithStore(pool).
		Build()

	server := api.
		Builder().
		WithStore(pool).
		WithConfig(config).
		WithStorage(storage).
		WithTokenGen(gen).
		Build()

	runDBMigration(config.MigrationLocation, config.ConnectionString)
	// Signal handler setup to catch SIGINT, SIGTERM
	srvSignalChan := server.SetupSignalHandler()
	prcSignalChan := processor.SetupSignalHandler()

	// Start background processor and server in goroutines
	go func() { srvChan <- processor.Work() }()
	go func() { prcChan <- server.Start() }()

	// Wait for shutdown signal or errors
	server.WaitForShutdown(srvChan, srvSignalChan)
	processor.WaitForShutdown(prcChan, prcSignalChan)
	log.Info().Msg("Application shut down gracefully.")
}

func runDBMigration(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create new migrate instance")
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Err(err).Msg("failed to run migrate up")
	}

	log.Info().Msg("db migrated successfully")
}
