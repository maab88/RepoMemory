package worker

import (
	"context"

	"github.com/maab88/repomemory/apps/worker/internal/config"
)

func Boot(ctx context.Context, cfg config.Config) error {
	return RunAsynqServer(ctx, cfg)
}
