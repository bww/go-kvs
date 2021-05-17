package redis

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/bww/go-kvs/v1"

	"github.com/stretchr/testify/assert"
)

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
}
