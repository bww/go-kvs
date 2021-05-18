package kvs

import (
	"context"
	"errors"
)

var (
	ErrNotFound    = errors.New("Not found")
	ErrInvalidDSN  = errors.New("Invalid DSN")
	ErrUnsupported = errors.New("Operation is not supported")
)

type Store interface {
	Get(cxt context.Context, key string, opts ...ReadOption) ([]byte, error)
	Set(cxt context.Context, key string, val []byte, opts ...WriteOption) error
	Inc(cxt context.Context, key string, inc int64, opts ...WriteOption) (int64, error)
	Delete(cxt context.Context, key string, opts ...WriteOption) error
}
