package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/maab88/repomemory/apps/api/internal/auth"
	"github.com/maab88/repomemory/apps/api/internal/config"
	"github.com/maab88/repomemory/apps/api/internal/db"
	gh "github.com/maab88/repomemory/apps/api/internal/github"
	"github.com/maab88/repomemory/apps/api/internal/http/handlers"
	"github.com/maab88/repomemory/apps/api/internal/http/router"
	jobdefs "github.com/maab88/repomemory/apps/api/internal/jobs"
	"github.com/maab88/repomemory/apps/api/internal/org"
	"github.com/maab88/repomemory/apps/api/internal/repositories"
	"github.com/maab88/repomemory/apps/api/internal/server"
	servicejobs "github.com/maab88/repomemory/apps/api/internal/services/jobs"
	servicememory "github.com/maab88/repomemory/apps/api/internal/services/memory"
	servicerepositories "github.com/maab88/repomemory/apps/api/internal/services/repositories"
	servicesearch "github.com/maab88/repomemory/apps/api/internal/services/search"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	_ = godotenv.Load(".env", "apps/api/.env")

	cfg := config.Load()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create database pool")
	}
	defer pool.Close()

	queries := db.New(pool)
	userResolver := auth.NewMockUserResolver(queries)
	orgStore := org.NewStore(pool, queries)
	orgService := org.NewService(orgStore)
	githubStore := gh.NewAccountRepository(queries)
	githubState := gh.NewMemoryStateService(cfg.GitHubStateSecret, cfg.GitHubStateTTL)
	githubClient := gh.NewHTTPGitHubClient(cfg.GitHubClientID, cfg.GitHubClientSecret, cfg.GitHubTokenURL, cfg.GitHubAPIBaseURL)
	githubOAuth := gh.NewOAuthService(gh.OAuthConfig{
		ClientID:     cfg.GitHubClientID,
		ClientSecret: cfg.GitHubClientSecret,
		AuthorizeURL: cfg.GitHubAuthorizeURL,
		TokenURL:     cfg.GitHubTokenURL,
		APIBaseURL:   cfg.GitHubAPIBaseURL,
		RedirectURL:  cfg.GitHubRedirectURL,
		StateSecret:  cfg.GitHubStateSecret,
		StateTTL:     cfg.GitHubStateTTL,
		Scope:        cfg.GitHubOAuthScope,
	}, githubState, githubClient, githubStore, gh.PlaintextTokenSealer{})
	githubRepositories := gh.NewRepositoryService(githubClient, githubStore)
	githubService := gh.NewService(githubOAuth, githubRepositories)
	jobRepository := repositories.NewJobRepository(queries)
	repositoryRepository := repositories.NewRepositoryRepository(queries)
	syncStateRepository := repositories.NewRepositorySyncStateRepository(queries)
	digestRepository := repositories.NewDigestRepository(queries)
	memoryEntryRepository := repositories.NewMemoryEntryRepository(queries)
	memoryEntrySourceRepository := repositories.NewMemoryEntrySourceRepository(queries)
	memorySearchRepository := repositories.NewMemorySearchRepository(queries)

	redisAddr := cfg.RedisAddr
	if !strings.Contains(redisAddr, "://") {
		redisAddr = "redis://" + redisAddr
	}
	redisOpt, err := asynq.ParseRedisURI(redisAddr)
	if err != nil {
		log.Fatal().Err(err).Msg("invalid redis address")
	}
	asynqClient := asynq.NewClient(redisOpt)
	defer asynqClient.Close()

	enqueuer := jobdefs.NewEnqueuer(jobRepository, asynqClient)
	jobService := servicejobs.NewService(jobRepository, queries, queries, enqueuer)
	repositoryService := servicerepositories.NewService(repositoryRepository, syncStateRepository, digestRepository, queries, jobService)
	memoryService := servicememory.NewQueryService(repositoryRepository, memoryEntryRepository, memoryEntrySourceRepository, queries)
	searchService := servicesearch.NewService(queries, repositoryRepository, memorySearchRepository)
	v1Handler := handlers.NewV1Handler(orgService, githubService, jobService, repositoryService, memoryService, searchService)

	h := router.New(router.Dependencies{
		AuthMiddleware: auth.RequireMockAuth(userResolver),
		V1Handler:      v1Handler,
	})

	srv := server.New(cfg, h)

	go func() {
		log.Info().Str("addr", srv.Addr).Msg("api starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("api failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("api shutdown failed")
	}

	log.Info().Msg("api stopped")
}
