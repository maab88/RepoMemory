package main

import (
	"github.com/maab88/repomemory/apps/worker/internal/config"
	"github.com/maab88/repomemory/apps/worker/internal/worker"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg := config.Load()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	log.Info().Str("env", cfg.Env).Str("redis", cfg.RedisAddr).Msg("worker booting")

	if err := worker.Boot(cfg); err != nil {
		log.Fatal().Err(err).Msg("worker failed to boot")
	}

	log.Info().Msg("worker boot completed")
}
