package worker

import (
	"context"
	"testing"

	"github.com/maab88/repomemory/apps/worker/internal/config"
)

func TestBoot(t *testing.T) {
	t.Run("rejects invalid uri", func(t *testing.T) {
		cfg := config.Config{RedisAddr: "://bad"}
		if err := Boot(context.Background(), cfg); err == nil {
			t.Fatal("expected error for invalid redis config")
		}
	})
}
