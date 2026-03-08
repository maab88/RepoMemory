package config

import "os"

type Config struct {
	Env           string
	RedisAddr     string
	DatabaseURL   string
	GitHubAPIBase string
	AIProvider    string
	OpenAIAPIKey  string
	OpenAIBaseURL string
	OpenAIModel   string
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

	aiProvider := os.Getenv("AI_PROVIDER")
	if aiProvider == "" {
		aiProvider = "disabled"
	}

	openAIBaseURL := os.Getenv("OPENAI_BASE_URL")
	if openAIBaseURL == "" {
		openAIBaseURL = "https://api.openai.com/v1/chat/completions"
	}

	openAIModel := os.Getenv("OPENAI_MODEL")
	if openAIModel == "" {
		openAIModel = "gpt-4o-mini"
	}

	return Config{
		Env:           env,
		RedisAddr:     redisAddr,
		DatabaseURL:   databaseURL,
		GitHubAPIBase: gitHubAPIBase,
		AIProvider:    aiProvider,
		OpenAIAPIKey:  os.Getenv("OPENAI_API_KEY"),
		OpenAIBaseURL: openAIBaseURL,
		OpenAIModel:   openAIModel,
	}
}
