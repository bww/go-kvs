package redis

type Config struct {
	Addr     string
	Password string
	Database int
}

func (c Config) WithOptions(opts ...Option) Config {
	for _, opt := range opts {
		c = opt(c)
	}
	return c
}

type Option func(Config) Config

func WithAddr(addr string) Option {
	return func(c Config) Config {
		c.Addr = addr
		return c
	}
}

func WithPassword(pass string) Option {
	return func(c Config) Config {
		c.Password = pass
		return c
	}
}

func WithDatabase(db int) Option {
	return func(c Config) Config {
		c.Database = db
		return c
	}
}
