package api

import (
	"context"

	db "github.com/FileConversionApi/db/sqlc"
	"github.com/FileConversionApi/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ServerBuilder struct {
	server *Server
}

func Builder() *ServerBuilder {
	builder := &ServerBuilder{&Server{}}
	return builder
}

func (builder *ServerBuilder) WithConfig(cfg utils.Config) *ServerBuilder {
	builder.server.cfg = cfg
	return builder
}
func (builder *ServerBuilder) WithTokenGen(gen utils.Generator) *ServerBuilder {
	builder.server.tokenGenerator = gen
	return builder
}
func (builder *ServerBuilder) WithStore(pool *pgxpool.Pool) *ServerBuilder {
	srvCtx, srvCancel := context.WithCancel(context.Background()) // Create context for the processor

	srvStore := db.NewStore(pool)

	builder.server.ctx = srvCtx
	builder.server.cancel = srvCancel
	builder.server.store = srvStore
	return builder
}

func (builder *ServerBuilder) WithStorage(storage utils.Storage) *ServerBuilder {
	builder.server.storage = storage
	return builder
}

func (builder *ServerBuilder) Build() *Server {
	builder.server.setupRouter()
	return builder.server
}
