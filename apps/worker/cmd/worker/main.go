package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/maab88/repomemory/apps/worker/internal/config"
	"github.com/maab88/repomemory/apps/worker/internal/worker"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	_ = godotenv.Load(".env", "apps/worker/.env")

	cfg := config.Load()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	log.Info().Str("env", cfg.Env).Str("redis", cfg.RedisAddr).Msg("worker booting")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := worker.Boot(ctx, cfg); err != nil {
		log.Fatal().Err(err).Msg("worker failed to boot")
	}

	log.Info().Msg("worker stopped")
}
