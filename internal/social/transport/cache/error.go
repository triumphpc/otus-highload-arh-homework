package cachewarmer

import "errors"

var (
	ErrCacheMiss    = errors.New("cache miss")
	ErrInvalidValue = errors.New("invalid value")
)
