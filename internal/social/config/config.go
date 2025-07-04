package config

import (
	"log"
	"time"

	cachewarmer "otus-highload-arh-homework/internal/social/transport/cache"
	"otus-highload-arh-homework/pkg/clients/kafka"
	"otus-highload-arh-homework/pkg/clients/pg"
	"otus-highload-arh-homework/pkg/clients/redis"

	"github.com/caarlos0/env/v9"
	"github.com/joho/godotenv"
)

type Config struct {
	HTTP  HTTP
	WS    WS
	App   App
	Auth  Auth
	Redis redis.Config
	PG    pg.Config
	Cache cachewarmer.Config
	Kafka kafka.Config
}

type HTTP struct {
	Port            string        `env:"HTTP_PORT" env-default:"8080"`
	ReadTimeout     time.Duration `env:"HTTP_READ_TIMEOUT" env-default:"5s"`
	WriteTimeout    time.Duration `env:"HTTP_WRITE_TIMEOUT" env-default:"5s"`
	IdleTimeout     time.Duration `env:"HTTP_IDLE_TIMEOUT" env-default:"30s"`
	ShutdownTimeout time.Duration `env:"HTTP_SHUTDOWN_TIMEOUT" env-default:"10s"`
}

type WS struct {
	Port string `env:"WS_PORT" env-default:"8081"`
}

type App struct {
	Env             string        `env:"APP_ENV" env-default:"development"`
	Debug           bool          `env:"APP_DEBUG" env-default:"false"`
	ShutdownTimeout time.Duration `env:"APP_SHUTDOWN_TIMEOUT" env-default:"10s"`
}

type Auth struct {
	HashCost     int           `env:"AUTH_HASH_COST" env-default:"10"`
	JwtSecretKey string        `env:"AUTH_JWT_SECRET_KEY" env-default:"your-secret-key"`
	JwtDuration  time.Duration `env:"AUTH_JWT_DURATION" env-default:"60m"`
}

func Load() *Config {
	_ = godotenv.Load()

	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	return &cfg
}
