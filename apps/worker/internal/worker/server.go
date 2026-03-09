package worker

import (
	"context"
	"fmt"
	"strings"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maab88/repomemory/apps/worker/internal/config"
	"github.com/maab88/repomemory/apps/worker/internal/jobs"
	"github.com/maab88/repomemory/apps/worker/internal/jobs/handlers"
	"github.com/maab88/repomemory/apps/worker/internal/services"
	workersai "github.com/maab88/repomemory/apps/worker/internal/services/ai"
	"github.com/maab88/repomemory/apps/worker/internal/services/hotspots"
	"github.com/rs/zerolog/log"
)

func RunAsynqServer(ctx context.Context, cfg config.Config) error {
	redisURI := cfg.RedisAddr
	if !strings.Contains(redisURI, "://") {
		redisURI = "redis://" + redisURI
	}
	redisOpt, err := asynq.ParseRedisURI(redisURI)
	if err != nil {
		return fmt.Errorf("parse redis config: %w", err)
	}

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("create db pool: %w", err)
	}
	defer pool.Close()

	store := jobs.NewStore(pool)
	githubSyncClient := services.NewHTTPGitHubSyncClient(cfg.GitHubAPIBase)
	initialSyncService := services.NewGitHubSyncService(store, githubSyncClient)
	aiProvider, err := workersai.NewProvider(workersai.Config{
		Provider:    cfg.AIProvider,
		OpenAIKey:   cfg.OpenAIAPIKey,
		OpenAIURL:   cfg.OpenAIBaseURL,
		OpenAIModel: cfg.OpenAIModel,
	})
	if err != nil {
		return fmt.Errorf("initialize ai provider: %w", err)
	}
	log.Info().Str("provider", aiProvider.Name()).Msg("worker ai provider configured")

	memoryGenerator := services.NewAIMemoryGenerator(aiProvider, services.NewDeterministicMemoryGenerator())
	memoryGenerationService := services.NewMemoryGenerationService(store, memoryGenerator)
	digestBuilder := services.NewAIDigestGenerator(aiProvider, services.NewDeterministicDigestBuilder())
	digestGenerationService := services.NewDigestGenerationService(store, digestBuilder)
	hotspotService := hotspots.NewService(store)
	initialSyncHandler := handlers.NewRepoInitialSyncHandler(store, initialSyncService)
	incrementalHandler := handlers.NewRepoIncrementalSyncHandler()
	generateMemoryHandler := handlers.NewGenerateMemoryHandler(store, memoryGenerationService)
	generateDigestHandler := handlers.NewGenerateDigestHandler(store, digestGenerationService)
	hotspotsHandler := handlers.NewRecalculateHotspotsHandler(store, hotspotService)

	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				jobs.QueueDefault: 1,
			},
			RetryDelayFunc: asynq.DefaultRetryDelayFunc,
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(jobs.TaskRepoInitialSync, initialSyncHandler.Handle)
	mux.HandleFunc(jobs.TaskRepoIncrementalSync, incrementalHandler.Handle)
	mux.HandleFunc(jobs.TaskRepoGenerateMemory, generateMemoryHandler.Handle)
	mux.HandleFunc(jobs.TaskRepoGenerateDigest, generateDigestHandler.Handle)
	mux.HandleFunc(jobs.TaskRepoRecalculateHotspots, hotspotsHandler.Handle)

	log.Info().Msg("worker asynq server starting")
	if err := server.Start(mux); err != nil {
		return fmt.Errorf("start asynq server: %w", err)
	}

	<-ctx.Done()
	log.Info().Msg("worker shutting down")
	server.Shutdown()

	return nil
}
