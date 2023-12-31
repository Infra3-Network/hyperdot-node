package apis

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"infra-3.xyz/hyperdot-node/internal/clients"
	"infra-3.xyz/hyperdot-node/internal/dataengine"

	"infra-3.xyz/hyperdot-node/internal/common"
	"infra-3.xyz/hyperdot-node/internal/store"
)

// ApiServer is a server to provide apis
type ApiServer struct {
	cfg    common.Config
	engine *gin.Engine
	srv    *http.Server
}

// NewApiServer creates a new ApiServer
func NewApiServer(boltStore *store.BoltStore, cfg *common.Config,
	db *gorm.DB,
	engines map[string]dataengine.QueryEngine, s3Client *clients.SimpleS3Cliet) (*ApiServer, error) {
	engine, err := NewRouterBuilder(boltStore, cfg, db, engines, s3Client).Build()
	if err != nil {
		return nil, err
	}

	return &ApiServer{
		cfg:    *cfg,
		engine: engine,
		srv: &http.Server{
			Addr:    cfg.ApiServer.Addr,
			Handler: engine,
		},
	}, nil
}

// Start starts the api server
func (s *ApiServer) GetEngine() *gin.Engine {
	return s.engine
}

// Start starts the api server
func (s *ApiServer) Start() error {
	go func() {
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
		return err
	}
	log.Println("Server exiting")
	return nil
}
