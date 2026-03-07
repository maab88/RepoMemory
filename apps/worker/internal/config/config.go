package config

import "os"

type Config struct {
	Env       string
	RedisAddr string
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

	return Config{Env: env, RedisAddr: redisAddr}
}
