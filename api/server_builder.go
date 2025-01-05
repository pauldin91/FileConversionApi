package api

import (
	"context"

	db "github.com/FileConversionApi/db/sqlc"
	"github.com/FileConversionApi/utils"
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
func (builder *ServerBuilder) WithStore(store db.Store) *ServerBuilder {
	builder.server.store = store
	return builder
}

func (builder *ServerBuilder) WithStorage(storage utils.Storage) *ServerBuilder {
	builder.server.storage = storage
	return builder
}
func (builder *ServerBuilder) WithCtx(parentCtx context.Context) *ServerBuilder {
	ctx, cancel := context.WithCancel(parentCtx)
	builder.server.ctx = ctx
	builder.server.cancel = cancel
	return builder
}

func (builder *ServerBuilder) Build() *Server {
	builder.server.setupRouter()
	return builder.server
}
