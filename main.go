package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/golang-migrate/migrate/v4"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	connPool, err := pgxpool.New(ctx, config.ConnectionString)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}

	runDBMigration(config.MigrationLocation, config.ConnectionString)

	store := db.NewStore(connPool)

	converter := utils.PdfConverter{}
	storage := utils.LocalStorage{}
	errChan := make(chan error, 1)

	// Launch the background processor
	go launchProcessor(store, storage, ctx, converter, errChan)

	// Set up signal handling for graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// Create a wait group to ensure all goroutines finish before exiting
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		select {
		case sig := <-signalChan:
			log.Info().Msgf("Received signal: %s. Shutting down...", sig)
			cancel()
		case err := <-errChan:
			log.Error().Err(err).Msg("Worker encountered an error. Shutting down...")
			cancel()
		}
	}()

	gen := utils.NewJwtGenerator(config.SigningKey)
	server := api.NewServer(config, gen, store, ctx, storage)

	go func() {
		if err := server.Start(); err != nil {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
		cancel()
	}()

	wg.Wait()
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

func launchProcessor(store db.Store, storage utils.Storage, ctx context.Context, converter utils.Converter, errChan chan error) {
	reqCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	processor := workers.NewDocumentProcessor(store, reqCtx, converter, storage)
	processor.Work(errChan)
}
