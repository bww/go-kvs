package redis

import (
	"context"

	"github.com/bww/go-kvs/v1"

	"github.com/bww/go-util/v1/errors"
	"github.com/go-redis/redis/v8"
)

type Store struct {
	*redis.Client
}

func New(opts ...Option) (*Store, error) {
	return NewWithConfig(Config{Addr: "localhost:6379"}.WithOptions(opts...))
}

func NewWithConfig(conf Config) (*Store, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       conf.Database,
	})
	return &Store{rdb}, nil
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
