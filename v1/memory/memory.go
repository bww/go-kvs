package memory

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/bww/go-kvs/v1"
	"github.com/hashicorp/golang-lru/v2/expirable"

	"github.com/bww/go-util/v1/errors"
)

const Scheme = "memory"

const maxAttempts = 10

type Store struct {
	Config
	sync.Mutex // mutex is only used on writes where bookkeeping updates are required; reads rely on the cache to manage synchronization
	cache      *expirable.LRU[string, []byte]
	nbytes     uint64
}

func New(dsn string, opts ...Option) (*Store, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}
	if u.Scheme != Scheme {
		return nil, fmt.Errorf("%w: expected scheme: %s (got: '%s' in '%s')", kvs.ErrInvalidDSN, Scheme, u.Scheme, dsn)
	}
	return NewWithConfig(Config{}.WithOptions(opts...))
}

func NewWithConfig(conf Config) (*Store, error) {
	s := &Store{Config: conf}
	s.cache = expirable.NewLRU[string, []byte](int(conf.Keys), s.evict, conf.TTL)
	return s, nil
}

// evict is called by LRU when an element is evict. We updated our
// bookkeeping to free its' associated capacity.
func (s *Store) evict(_ string, v []byte) {
	s.Lock()
	s.nbytes -= uint64(len(v))
	s.Unlock()
}

// alloc determines if there is sufficient alloc to store the specified
// number of bytes. If so, it increments the consumed alloc and returns true.
// If not, it returns false.
func (s *Store) alloc(n uint64) (res bool) {
	s.Lock()
	res = s.Bytes <= 0 || (s.nbytes+n) <= s.Bytes
	if res {
		s.nbytes += n
	}
	s.Unlock()
	return
}

func (s *Store) Len() int64 {
	return int64(s.cache.Len())
}

func (s *Store) Cap() (uint64, uint64) {
	if s.Bytes <= 0 {
		return 0, 0 // capacity is not limited; return zeros
	}
	s.Lock()
	n := s.nbytes
	s.Unlock()
	return s.Bytes, n
}

func (s *Store) Keys(cxt context.Context, opts ...kvs.ReadOption) (kvs.Iter[string], error) {
	conf := kvs.ReadConfig{}.WithOptions(opts...)
	keys := make(chan string, 100)
	iter := keysIter{data: keys}
	go func() {
		for _, e := range s.cache.Keys() {
			if conf.Prefix == "" || strings.HasPrefix(e, conf.Prefix) {
				keys <- e
			}
		}
		close(keys)
	}()
	return iter, nil
}

func (s *Store) Get(cxt context.Context, key string, opts ...kvs.ReadOption) ([]byte, error) {
	val, ok := s.cache.Get(key)
	if !ok {
		return nil, errors.Wrapf(kvs.ErrNotFound, "Not found: %s", key)
	}
	return val, nil
}

func (s *Store) Set(cxt context.Context, key string, val []byte, opts ...kvs.WriteOption) error {
	// If we have a memory consumption limit, make a limitied number of attempts
	// to free enough capacity for the insertion. If we cannot do so, the
	// insertion fails.
	if s.Bytes > 0 {
		for i := 0; ; i++ {
			if !s.alloc(uint64(len(val))) {
				if i < maxAttempts && s.cache.Len() > 0 {
					s.cache.RemoveOldest()
					continue // remove the oldest element and try again...
				} else {
					return kvs.ErrCapacityExceeded
				}
			}
			break // ok, we have room
		}
	}
	s.cache.Add(key, val)
	return nil
}

func (s *Store) Inc(cxt context.Context, key string, inc int64, opts ...kvs.WriteOption) (int64, error) {
	return -1, kvs.ErrNotSupported
}

func (s *Store) Delete(cxt context.Context, key string, opts ...kvs.WriteOption) error {
	// If the element exists in the LRU, the eviction handler is called to deal
	// with freeing capacity. If the element does not exist in the LRU, this
	// operation has no effect.
	s.cache.Remove(key)
	return nil
}