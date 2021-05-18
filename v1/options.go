package kvs

import (
	"time"
)

type ReadConfig struct{}

func (c ReadConfig) WithOptions(opts ...ReadOption) ReadConfig {
	for _, opt := range opts {
		c = opt(c)
	}
	return c
}

type ReadOption func(ReadConfig) ReadConfig

type WriteConfig struct {
	TTL        time.Duration
	Expiration time.Time
}

func (c WriteConfig) WithOptions(opts ...WriteOption) WriteConfig {
	for _, opt := range opts {
		c = opt(c)
	}
	return c
}

type WriteOption func(WriteConfig) WriteConfig

func WithTTL(ttl time.Duration) WriteOption {
	return func(c WriteConfig) WriteConfig {
		c.TTL = ttl
		return c
	}
}

func WithExpiration(when time.Time) WriteOption {
	return func(c WriteConfig) WriteConfig {
		c.Expiration = when
		return c
	}
}
