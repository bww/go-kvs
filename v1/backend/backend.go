package backend

import (
	"errors"
	"strings"

	"github.com/bww/go-kvs/v1"
	"github.com/bww/go-kvs/v1/backend/memory"
	"github.com/bww/go-kvs/v1/backend/redis"
)

var errUnsupported = errors.New("Service not supported")

func New(dsn string) (kvs.Store, error) {
	var scheme string
	if x := strings.Index(dsn, ":"); x > 0 {
		scheme = dsn[:x]
	} else {
		scheme = dsn
	}
	switch scheme {
	case memory.Scheme:
		return memory.New(dsn)
	case redis.Scheme:
		return redis.New(dsn)
	default:
		return nil, errUnsupported
	}
}
