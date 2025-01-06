package api

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	db "github.com/FileConversionApi/db/sqlc"
	"github.com/FileConversionApi/utils"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
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

func (server *Server) WaitForShutdown(errChan chan error, signalChan chan os.Signal) {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		select {
		case sig := <-signalChan:
			log.Info().Msgf("Received signal: %s. Shutting down...", sig)
			server.cancel()
		case <-errChan:
			for {
				err := <-errChan
				if err != nil {
					log.Info().Msgf("Received signal: %s. Shutting down...", err)
					server.cancel()
				}
			}
		}

	}()

	wg.Wait()
}

func (server *Server) SetupSignalHandler() chan os.Signal {
	signalChan := make(chan os.Signal, 1) // Buffered channel to avoid blocking

	// Notify the channel on interrupt (SIGINT) and terminate (SIGTERM) signals
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Goroutine to handle the signal and trigger cancellations
	go func() {
		sig := <-signalChan
		log.Info().Msgf("Received signal: %s. Shutting down...", sig)
		server.cancel()

	}()

	return signalChan
}
