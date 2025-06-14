package redis

import "time"

type Config struct {
	Address      string        `env:"REDIS_ADDRESS" env-default:"localhost:6379"`
	Password     string        `env:"REDIS_PASSWORD" env-default:""`
	DB           int           `env:"REDIS_DB" env-default:"0"`
	PoolSize     int           `env:"REDIS_POOL_SIZE" env-default:"20"`
	DialTimeout  time.Duration `env:"REDIS_DIAL_TIMEOUT" env-default:"5s"`
	ReadTimeout  time.Duration `env:"REDIS_READ_TIMEOUT" env-default:"3s"`
	WriteTimeout time.Duration `env:"REDIS_WRITE_TIMEOUT" env-default:"3s"`
	MaxRetries   int           `env:"REDIS_MAX_RETRIES" env-default:"3"`
}
