package pg

import (
	"time"
)

type Config struct {
	URL               string        `env:"PG_URL,required"`
	MaxOpenConns      int           `env:"PG_MAX_OPEN_CONNS" env-default:"10"`
	MaxIdleConns      int           `env:"PG_MAX_IDLE_CONNS" env-default:"5"`
	ConnMaxLifetime   time.Duration `env:"PG_CONN_MAX_LIFETIME" env-default:"5m"`
	ConnMaxIdleTime   time.Duration `env:"PG_CONN_MAX_IDLE_TIME" env-default:"1m"`
	HealthCheckPeriod time.Duration `env:"PG_HEALTH_CHECK_PERIOD" env-default:"30s"`
}
