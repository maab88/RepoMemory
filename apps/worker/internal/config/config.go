package config

import "os"

type Config struct {
	Env           string
	RedisAddr     string
	DatabaseURL   string
	GitHubAPIBase string
}

func Load() Config {
	env := os.Getenv("WORKER_ENV")
	if env == "" {
		env = "development"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "127.0.0.1:6379"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/repomemory?sslmode=disable"
	}

	gitHubAPIBase := os.Getenv("GITHUB_API_BASE_URL")
	if gitHubAPIBase == "" {
		gitHubAPIBase = "https://api.github.com"
	}

	return Config{Env: env, RedisAddr: redisAddr, DatabaseURL: databaseURL, GitHubAPIBase: gitHubAPIBase}
}
