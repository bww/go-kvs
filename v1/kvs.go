package kvs

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrNotFound         = errors.New("Not found")
	ErrInvalidDSN       = errors.New("Invalid DSN")
	ErrNotSupported     = errors.New("Not supported")
	ErrClosed           = errors.New("Result set closed")
	ErrCapacityExceeded = errors.New("Exceeded capacity")
)

type Iter[T any] interface {
	Next() (T, error)
}

type Store interface {
	fmt.Stringer
	Keys(cxt context.Context, opts ...ReadOption) (Iter[string], error)
	Get(cxt context.Context, key string, opts ...ReadOption) ([]byte, error)
	Set(cxt context.Context, key string, val []byte, opts ...WriteOption) error
	Inc(cxt context.Context, key string, inc int64, opts ...WriteOption) (int64, error)
	Delete(cxt context.Context, key string, opts ...WriteOption) error
}
