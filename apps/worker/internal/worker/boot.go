package worker

import (
	"fmt"
	"strings"

	"github.com/hibiken/asynq"
	"github.com/maab88/repomemory/apps/worker/internal/config"
)

func Boot(cfg config.Config) error {
	redisURI := cfg.RedisAddr
	if !strings.Contains(redisURI, "://") {
		redisURI = "redis://" + redisURI
	}

	if _, err := asynq.ParseRedisURI(redisURI); err != nil {
		return fmt.Errorf("parse redis config: %w", err)
	}
	return nil
}