package config

import (
	"os"
	"time"
)

type Config struct {
	Port               string
	Env                string
	AuthMode           string
	AuthJWTSecret      string
	AuthJWTIssuer      string
	AuthJWTAudience    string
	DatabaseURL        string
	RedisAddr          string
	GitHubClientID     string
	GitHubClientSecret string
	GitHubAuthorizeURL string
	GitHubTokenURL     string
	GitHubAPIBaseURL   string
	GitHubRedirectURL  string
	GitHubStateSecret  string
	GitHubStateTTL     time.Duration
	GitHubOAuthScope   string
}

func Load() Config {
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("API_ENV")
	if env == "" {
		env = "development"
	}
	authMode := os.Getenv("API_AUTH_MODE")
	if authMode == "" {
		authMode = "jwt"
	}
	authJWTIssuer := os.Getenv("API_AUTH_JWT_ISSUER")
	if authJWTIssuer == "" {
		authJWTIssuer = "repomemory-web"
	}
	authJWTAudience := os.Getenv("API_AUTH_JWT_AUDIENCE")
	if authJWTAudience == "" {
		authJWTAudience = "repomemory-api"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/repomemory?sslmode=disable"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "127.0.0.1:6379"
	}

	gitHubAuthorizeURL := os.Getenv("GITHUB_AUTHORIZE_URL")
	if gitHubAuthorizeURL == "" {
		gitHubAuthorizeURL = "https://github.com/login/oauth/authorize"
	}

	gitHubTokenURL := os.Getenv("GITHUB_TOKEN_URL")
	if gitHubTokenURL == "" {
		gitHubTokenURL = "https://github.com/login/oauth/access_token"
	}

	gitHubAPIBaseURL := os.Getenv("GITHUB_API_BASE_URL")
	if gitHubAPIBaseURL == "" {
		gitHubAPIBaseURL = "https://api.github.com"
	}

	gitHubRedirectURL := os.Getenv("GITHUB_REDIRECT_URL")
	if gitHubRedirectURL == "" {
		gitHubRedirectURL = "http://localhost:3000/integrations/github/callback"
	}

	gitHubStateTTL := 10 * time.Minute
	gitHubStateTTLRaw := os.Getenv("GITHUB_STATE_TTL")
	if gitHubStateTTLRaw != "" {
		if parsed, err := time.ParseDuration(gitHubStateTTLRaw); err == nil && parsed > 0 {
			gitHubStateTTL = parsed
		}
	}

	gitHubOAuthScope := os.Getenv("GITHUB_OAUTH_SCOPE")
	if gitHubOAuthScope == "" {
		gitHubOAuthScope = "repo read:user user:email"
	}

	return Config{
		Port:               port,
		Env:                env,
		AuthMode:           authMode,
		AuthJWTSecret:      os.Getenv("API_AUTH_JWT_SECRET"),
		AuthJWTIssuer:      authJWTIssuer,
		AuthJWTAudience:    authJWTAudience,
		DatabaseURL:        databaseURL,
		RedisAddr:          redisAddr,
		GitHubClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		GitHubClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		GitHubAuthorizeURL: gitHubAuthorizeURL,
		GitHubTokenURL:     gitHubTokenURL,
		GitHubAPIBaseURL:   gitHubAPIBaseURL,
		GitHubRedirectURL:  gitHubRedirectURL,
		GitHubStateSecret:  os.Getenv("GITHUB_STATE_SECRET"),
		GitHubStateTTL:     gitHubStateTTL,
		GitHubOAuthScope:   gitHubOAuthScope,
	}
}
