package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/maab88/repomemory/apps/api/internal/config"
	"github.com/maab88/repomemory/apps/api/internal/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg := config.Load()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	srv := server.New(cfg)

	go func() {
		log.Info().Str("addr", srv.Addr).Msg("api starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("api failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("api shutdown failed")
	}

	log.Info().Msg("api stopped")
}
