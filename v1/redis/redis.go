package redis

import (
	"context"
	"net/url"
	"path"
	"strconv"

	"github.com/bww/go-kvs/v1"

	"github.com/bww/go-util/v1/errors"
	"github.com/go-redis/redis/v8"
)

const Scheme = "redis"

type Store struct {
	*redis.Client
}

func New(dsn string, opts ...Option) (*Store, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}
	if u.Scheme != Scheme {
		return nil, kvs.ErrInvalidDSN
	}
	var db int
	if p := u.Path; p != "" && p != "/" {
		db, err = strconv.Atoi(path.Base(p))
		if err != nil {
			return nil, err
		}
	}
	return NewWithConfig(Config{Addr: u.Host, Database: db}.WithOptions(opts...))
}

func NewWithConfig(conf Config) (*Store, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       conf.Database,
	})
	return &Store{rdb}, nil
}

func (s *Store) Keys(cxt context.Context, opts ...kvs.ReadOption) (kvs.Iter[string], error) {
	conf := kvs.ReadConfig{}.WithOptions(opts...)
	var it *redis.ScanIterator
	if pfx := conf.Prefix; pfx != "" {
		it = s.Client.Scan(cxt, 0, "prefix:"+pfx, 0).Iterator()
	} else {
		it = s.Client.Scan(cxt, 0, "", 0).Iterator()
	}
	return scanIter{
		cxt:  cxt,
		iter: it,
	}, nil
}

func (s *Store) Get(cxt context.Context, key string, opts ...kvs.ReadOption) ([]byte, error) {
	val, err := s.Client.Get(cxt, key).Result()
	if err == redis.Nil {
		return nil, errors.Wrapf(kvs.ErrNotFound, "Not found: %s", key)
	} else if err != nil {
		return nil, err
	}
	return []byte(val), nil
}

func (s *Store) Set(cxt context.Context, key string, val []byte, opts ...kvs.WriteOption) error {
	conf := kvs.WriteConfig{}.WithOptions(opts...)
	return s.Client.Set(cxt, key, string(val), conf.TTL).Err()
}

func (s *Store) Delete(cxt context.Context, key string, opts ...kvs.WriteOption) error {
	return s.Client.Del(cxt, key).Err()
}
