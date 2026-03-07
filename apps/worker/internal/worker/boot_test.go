package worker

import (
	"testing"

	"github.com/maab88/repomemory/apps/worker/internal/config"
)

func TestBoot(t *testing.T) {
	t.Run("accepts host and port", func(t *testing.T) {
		cfg := config.Config{RedisAddr: "127.0.0.1:6379"}
		if err := Boot(cfg); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("rejects invalid uri", func(t *testing.T) {
		cfg := config.Config{RedisAddr: "://bad"}
		if err := Boot(cfg); err == nil {
			t.Fatal("expected error for invalid redis config")
		}
	})
}
