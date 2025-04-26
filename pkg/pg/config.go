package pg

import (
	"log"
	"time"

	"github.com/caarlos0/env/v9"
	"github.com/joho/godotenv"
)

type Config struct {
	URL               string        `env:"PG_URL,required"`
	MaxOpenConns      int           `env:"PG_MAX_OPEN_CONNS" env-default:"10"`
	MaxIdleConns      int           `env:"PG_MAX_IDLE_CONNS" env-default:"5"`
	ConnMaxLifetime   time.Duration `env:"PG_CONN_MAX_LIFETIME" env-default:"5m"`
	ConnMaxIdleTime   time.Duration `env:"PG_CONN_MAX_IDLE_TIME" env-default:"1m"`
	HealthCheckPeriod time.Duration `env:"PG_HEALTH_CHECK_PERIOD" env-default:"30s"`
}

func Load() *Config {
	_ = godotenv.Load()

	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	return &cfg
}
