package main

import (
	"context"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/FileConversionApi/api"
	db "github.com/FileConversionApi/db/sqlc"
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

	connPool, err := pgxpool.New(context.Background(), config.ConnectionString)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}

	runDBMigration(config.MigrationLocation, config.ConnectionString)

	store := db.NewStore(connPool)

	// Start the HTTPS server
	gen := utils.NewJwtGenerator(config.SigningKey)
	server := api.NewServer(config, gen, store)
	err = server.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
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
