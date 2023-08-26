package apis

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"infra-3.xyz/hyperdot-node/internal/common"
)

type ApiServer struct {
	cfg common.Config
	srv *http.Server
}

func NewApiServer(cfg *common.Config) (*ApiServer, error) {
	engine, err := NewRouterBuilder(cfg).Build()
	if err != nil {
		return nil, err
	}
	return &ApiServer{
		cfg: *cfg,
		srv: &http.Server{
			Addr:    cfg.ApiServer.Addr,
			Handler: engine,
		},
	}, nil
}

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
