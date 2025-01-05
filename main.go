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
	signalChan := setupSignalHandler(cancel)

	connPool, err := pgxpool.New(ctx, config.ConnectionString)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}
	defer connPool.Close()

	runDBMigration(config.MigrationLocation, config.ConnectionString)

	store := db.NewStore(connPool)

	storage := utils.LocalStorage{}
	errChan := make(chan error, 2)

	converter := utils.PdfConverter{}

	processor := workers.Builder().
		WithConverter(converter).
		WithCtx(ctx).
		WithStorage(storage).
		WithStore(store).
		Build()

	gen := utils.NewJwtGenerator(config.SigningKey)
	server := api.
		Builder().
		WithCtx(ctx).
		WithStore(store).
		WithConfig(config).
		WithStorage(storage).
		WithTokenGen(gen).
		Build()

	go processor.Work(errChan)
	go server.Start()

	waitForShutdown(errChan, signalChan, cancel)

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

func waitForShutdown(errChan chan error, signalChan chan os.Signal, cancel context.CancelFunc) {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		// Wait for either a signal or an error from the worker/server
		select {
		case sig := <-signalChan:
			log.Info().Msgf("Received signal: %s. Shutting down...", sig)
			cancel() // Cancel the context to initiate shutdown
		case err := <-errChan:
			log.Error().Err(err).Msg("Worker/server encountered an error. Shutting down...")
			cancel() // Cancel the context to initiate shutdown
		}
	}()

	// Wait for the goroutine above to complete before allowing the app to exit
	wg.Wait()
}

func setupSignalHandler(cancel context.CancelFunc) chan os.Signal {
	signalChan := make(chan os.Signal, 1) // Buffered channel to avoid blocking

	// Notify the channel on interrupt (SIGINT) and terminate (SIGTERM) signals
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// This goroutine will trigger the cancel function when a signal is received
	go func() {
		sig := <-signalChan
		log.Info().Msgf("Received signal: %s. Shutting down...", sig)
		cancel() // Trigger cancellation of the context to initiate graceful shutdown
	}()

	return signalChan
}
