package api

import (
	"context"
	"path/filepath"

	db "github.com/FileConversionApi/db/sqlc"
	"github.com/FileConversionApi/utils"
	"github.com/gin-gonic/gin"
)

type Server struct {
	cfg            utils.Config
	store          db.Store
	router         *gin.Engine
	ctx            context.Context
	tokenGenerator utils.Generator
	storage        utils.Storage
	cancel         context.CancelFunc
}

func NewServer(cfg utils.Config, tokenGenerator utils.Generator, store db.Store, parentCtx context.Context, storage utils.Storage) *Server {
	ctx, cancel := context.WithCancel(parentCtx)
	server := &Server{
		cfg:            cfg,
		store:          store,
		tokenGenerator: tokenGenerator,
		ctx:            ctx,
		cancel:         cancel,
		storage:        storage,
	}

	server.setupRouter()

	return server
}

func (server *Server) Start() error {
	certFile := filepath.Join(server.cfg.CertPath, server.cfg.CertFile)
	certKey := filepath.Join(server.cfg.CertPath, server.cfg.CertKey)
	return server.router.RunTLS(server.cfg.HttpServerAddress, certFile, certKey)
}

func (server *Server) setupRouter() {

	router := gin.Default()

	//swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(server.statikFS))
	//router.GET("/swagger/*filepath", gin.WrapH(swaggerHandler))

	router.POST(usersRoute, server.createUser)
	router.POST("/auth", server.login)

	authRoutes := router.Group("/").Use(server.authorize())

	authRoutes.GET(usersRoute, server.listUsers)
	authRoutes.GET(usersRoute+"/:email", server.getUser)

	authRoutes.POST(documents, server.convert)
	authRoutes.GET(documents+"/:id", server.retrieve)
	server.router = router

}
