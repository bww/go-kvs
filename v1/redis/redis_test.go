package redis

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/bww/go-kvs/v1"

	"github.com/stretchr/testify/assert"
)

func resultSet[T comparable](v ...T) map[T]struct{} {
	r := make(map[T]struct{})
	for _, e := range v {
		r[e] = struct{}{}
	}
	return r
}

func TestRedisCRUD(t *testing.T) {
	var res []byte

	store, err := New("redis://localhost:59011/")
	if !assert.Nil(t, err, fmt.Sprint(err)) {
		return
	}

	err = store.Set(context.Background(), "a", []byte("This is A"))
	assert.Nil(t, err, fmt.Sprint(err))
	err = store.Set(context.Background(), "b", []byte("This is B"))
	assert.Nil(t, err, fmt.Sprint(err))

	res, err = store.Get(context.Background(), "a")
	if assert.Nil(t, err, fmt.Sprint(err)) {
		assert.Equal(t, []byte("This is A"), res)
	}
	res, err = store.Get(context.Background(), "b")
	if assert.Nil(t, err, fmt.Sprint(err)) {
		assert.Equal(t, []byte("This is B"), res)
	}

	err = store.Delete(context.Background(), "b")
	assert.Nil(t, err, fmt.Sprint(err))
	_, err = store.Get(context.Background(), "b")
	assert.Equal(t, true, errors.Is(err, kvs.ErrNotFound))

	err = store.Delete(context.Background(), "a")
	assert.Nil(t, err, fmt.Sprint(err))
}

func TestRedisKeys(t *testing.T) {
	var expect map[string]struct{}

	store, err := New("redis://localhost:59011/")
	if !assert.Nil(t, err, fmt.Sprint(err)) {
		return
	}

	allkeys := []string{"a.1", "a.2", "b.1", "b.2"}
	for _, e := range allkeys {
		err = store.Set(context.Background(), e, []byte("This is the value for "+e))
		assert.Nil(t, err, fmt.Sprint(err))
	}

	expect = resultSet(allkeys...)
	kit, err := store.Keys(context.Background())
	for {
		val, err := kit.Next()
		if err == kvs.ErrClosed {
			break
		}
		_, ok := expect[val]
		assert.Equal(t, true, ok)
		delete(expect, val)
	}
	assert.Len(t, expect, 0)

	expect = resultSet("a.1", "a.2")
	kit, err = store.Keys(context.Background(), kvs.WithPrefix("a."))
	for {
		val, err := kit.Next()
		if err == kvs.ErrClosed {
			break
		}
		_, ok := expect[val]
		assert.Equal(t, true, ok)
		delete(expect, val)
	}
	assert.Len(t, expect, 0)

	for _, e := range allkeys {
		err = store.Delete(context.Background(), e)
		assert.Nil(t, err, fmt.Sprint(err))
	}
}
