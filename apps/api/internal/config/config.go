package config

import "os"

type Config struct {
	Port        string
	Env         string
	DatabaseURL string
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

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/repomemory?sslmode=disable"
	}

	return Config{Port: port, Env: env, DatabaseURL: databaseURL}
}
