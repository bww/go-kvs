package memory

import "time"

type Config struct {
	MaxKeys  uint64 // maximum keys allowed
	MaxBytes uint64 // maximum bytes allowed
	TTL      time.Duration
}

func (c Config) WithOptions(opts ...Option) Config {
	for _, opt := range opts {
		c = opt(c)
	}
	return c
}

type Option func(Config) Config

func WithMaxKeys(n uint64) Option {
	return func(c Config) Config {
		c.MaxKeys = n
		return c
	}
}

func WithMaxBytes(n uint64) Option {
	return func(c Config) Config {
		c.MaxBytes = n
		return c
	}
}
