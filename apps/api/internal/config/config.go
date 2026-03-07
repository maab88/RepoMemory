package config

import "os"

type Config struct {
	Port string
	Env  string
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

	return Config{Port: port, Env: env}
}
