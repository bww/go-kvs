package kvs

import (
	"context"
	"errors"
)

var (
	ErrNotFound   = errors.New("Not found")
	ErrInvalidDSN = errors.New("Invalid DSN")
)

type Store interface {
	Get(cxt context.Context, key string, opts ...ReadOption) ([]byte, error)
	Set(cxt context.Context, key string, val []byte, opts ...WriteOption) error
	Delete(cxt context.Context, key string, opts ...WriteOption) error
}
