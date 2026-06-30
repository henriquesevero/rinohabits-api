package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"

	httpapi "github.com/henriquesevero/rinohabits-api/internal/adapter/http"
	"github.com/henriquesevero/rinohabits-api/internal/adapter/postgres"
	"github.com/henriquesevero/rinohabits-api/internal/adapter/security"
	"github.com/henriquesevero/rinohabits-api/internal/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	tokenManager := security.NewJWTTokenManager(cfg.JWTSecret, 7*24*time.Hour)

	server := &http.Server{
		Addr: ":" + cfg.Port,
		Handler: httpapi.NewRouter(httpapi.Dependencies{
			Pool:         pool,
			TokenManager: tokenManager,
			CORSOrigin:   cfg.CORSOrigin,
			UploadsDir:   cfg.UploadsDir,
			APIBaseURL:   cfg.APIBaseURL,
		}),
	}

	go func() {
		log.Printf("rinohabits-api listening on :%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("graceful shutdown failed: %v", err)
	}

	log.Println("rinohabits-api stopped")
}
