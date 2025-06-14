package cachewarmer

import "time"

type Config struct {
	Enabled         bool          `env:"CACHE_ENABLED" env-default:"true"`
	Type            string        `env:"CACHE_TYPE" env-default:"redis"` // redis or inmemory
	TTL             time.Duration `env:"CACHE_TTL" env-default:"24h"`
	CleanupInterval time.Duration `env:"CACHE_CLEANUP_INTERVAL" env-default:"1h"`
	Size            int           `env:"CACHE_SIZE" env-default:"10000"` // for inmemory cache
	NumWorkers      int           `env:"CACHE_NUM_WORKERS" env-default:"8"`
}
