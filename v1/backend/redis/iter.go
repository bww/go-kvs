package redis

import (
	"context"

	"github.com/bww/go-kvs/v1"

	"github.com/go-redis/redis/v8"
)

type scanIter struct {
	cxt  context.Context
	iter *redis.ScanIterator
}

func (r scanIter) Next() (string, error) {
	if r.iter.Next(r.cxt) {
		return r.iter.Val(), nil
	} else if err := r.iter.Err(); err != nil {
		return "", err
	} else {
		return "", kvs.ErrClosed
	}
}
