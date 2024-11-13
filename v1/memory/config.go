package memory

import "time"

type Config struct {
	Keys  uint64 // maximum keys allowed
	Bytes uint64 // maximum bytes allowed
	TTL   time.Duration
}

func (c Config) WithOptions(opts ...Option) Config {
	for _, opt := range opts {
		c = opt(c)
	}
	return c
}

type Option func(Config) Config

func WithKeys(n uint64) Option {
	return func(c Config) Config {
		c.Keys = n
		return c
	}
}

func WithBytes(n uint64) Option {
	return func(c Config) Config {
		c.Bytes = n
		return c
	}
}
