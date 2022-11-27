package kvs

import (
	"time"
)

type ReadConfig struct {
	Prefix string
}

func (c ReadConfig) WithOptions(opts ...ReadOption) ReadConfig {
	for _, opt := range opts {
		c = opt(c)
	}
	return c
}

type ReadOption func(ReadConfig) ReadConfig

func WithPrefix(pfx string) ReadOption {
	return func(c ReadConfig) ReadConfig {
		c.Prefix = pfx
		return c
	}
}

type WriteConfig struct {
	TTL time.Duration
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
