package memory

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

func incRune(b rune, n int) string {
	x := b + rune(n)
	return string(x)
}

func TestMemoryCRUD(t *testing.T) {
	var res []byte

	store, err := New("memory:")
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

	kit, err := store.Keys(context.Background())
	for {
		key, err := kit.Next()
		if err == kvs.ErrClosed {
			break
		}
		store.Delete(context.Background(), key)
	}
}

func TestMemoryEviction(t *testing.T) {
	var res []byte

	c := uint64(128)
	store, err := New("memory:", WithBytes(c))
	if !assert.Nil(t, err, fmt.Sprint(err)) {
		return
	}

	n, m := 8, uint64(16)
	val := []byte("This is 16 bytes")
	for i := 0; i < n; i++ {
		err = store.Set(context.Background(), incRune('a', i), val)
		assert.Nil(t, err, fmt.Sprint(err))
	}

	kit, err := store.Keys(context.Background())
	for {
		val, err := kit.Next()
		if err == kvs.ErrClosed {
			break
		}
		fmt.Println(">>> >>> >>>", val)
	}

	// All items have been added
	assert.Equal(t, int64(n), store.Len())
	// Capacity is full
	total, used := store.Cap()
	assert.Equal(t, c, total)
	assert.Equal(t, c, used)

	// Adding this will force an eviction for space
	err = store.Set(context.Background(), incRune('a', n+1), val)
	assert.Nil(t, err, fmt.Sprint(err))

	// We still have 8 items
	assert.Equal(t, int64(n), store.Len())
	// 'a' is gone
	res, err = store.Get(context.Background(), "a")
	assert.ErrorIs(t, err, kvs.ErrNotFound)
	// The last key is present
	res, err = store.Get(context.Background(), incRune('a', n+1))
	if assert.Nil(t, err, fmt.Sprint(err)) {
		assert.Equal(t, val, res)
	}

	// Explicitly delete a key
	err = store.Delete(context.Background(), "b")
	assert.Nil(t, err, fmt.Sprint(err))

	// All items have been added
	assert.Equal(t, int64(n-1), store.Len())
	// Capacity is made available
	total, used = store.Cap()
	assert.Equal(t, c, total)
	assert.Equal(t, c-m, used)

}

func TestMemoryKeys(t *testing.T) {
	var expect map[string]struct{}

	store, err := New("memory:")
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

	kit, err = store.Keys(context.Background())
	for {
		key, err := kit.Next()
		if err == kvs.ErrClosed {
			break
		}
		store.Delete(context.Background(), key)
	}
}
