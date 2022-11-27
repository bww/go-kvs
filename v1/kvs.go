package kvs

import (
	"context"
	"errors"
)

var (
	ErrNotFound     = errors.New("Not found")
	ErrInvalidDSN   = errors.New("Invalid DSN")
	ErrNotSupported = errors.New("Not supported")
	ErrClosed       = errors.New("Result set closed")
)

type Iter[T any] interface {
	Next() (T, error)
}

type Store interface {
	Keys(cxt context.Context, opts ...ReadOption) (Iter[string], error)
	Get(cxt context.Context, key string, opts ...ReadOption) ([]byte, error)
	Set(cxt context.Context, key string, val []byte, opts ...WriteOption) error
	Delete(cxt context.Context, key string, opts ...WriteOption) error
}
