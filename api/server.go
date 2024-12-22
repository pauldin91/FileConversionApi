package api

import (
	"path/filepath"

	db "github.com/FileConversionApi/db/sqlc"
	"github.com/FileConversionApi/utils"
	"github.com/gin-gonic/gin"
)

type Server struct {
	cfg            utils.Config
	store          db.Store
	router         *gin.Engine
	tokenGenerator utils.Generator
}

func NewServer(cfg utils.Config, tokenGenerator utils.Generator, store db.Store) *Server {

	server := &Server{
		cfg:            cfg,
		store:          store,
		tokenGenerator: tokenGenerator,
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

	router.POST("/users", server.createUser)
	router.POST("/auth", server.login)

	server.router = router

}
