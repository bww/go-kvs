package memory

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/bww/go-kvs/v1"

	"github.com/bww/go-util/v1/ext"
	"github.com/bww/golang-lru/v2/expirable"
	"github.com/dustin/go-humanize"
)

const Scheme = "memory"

const maxAttempts = 10

var errCompareFailed = errors.New("Comparison failed")

type Store struct {
	Config
	sync.Mutex // mutex is only used on writes where bookkeeping updates are required; reads rely on the cache to manage synchronization
	cache      *expirable.LRU[string, []byte]
	nbytes     uint64
}

func New(dsn string, opts ...Option) (*Store, error) {
	var keys int64
	var bytes uint64
	if strings.Index(dsn, ":") > 0 {
		u, err := url.Parse(dsn)
		if err != nil {
			return nil, fmt.Errorf("Invalid DSN: %w", err)
		}
		q := u.Query()
		if v := q.Get("keys"); v != "" {
			keys, err = strconv.ParseInt(v, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("Invalid key count: %w", err)
			}
		}
		if v := q.Get("bytes"); v != "" {
			bytes, err = humanize.ParseBytes(v)
			if err != nil {
				return nil, fmt.Errorf("Invalid storage size: %w", err)
			}
		}
	}
	return NewWithConfig(Config{
		MaxKeys:  uint64(keys),
		MaxBytes: bytes,
	}.WithOptions(opts...))
}

func NewWithConfig(conf Config) (*Store, error) {
	s := &Store{Config: conf}
	s.cache = expirable.NewLRU[string, []byte](int(conf.MaxKeys), s.evict, conf.TTL)
	return s, nil
}

func (s *Store) String() string {
	var sb strings.Builder
	var n int
	sb.WriteString("Memory")
	if s.MaxKeys > 0 {
		sb.WriteString(ext.Choose(n == 0, ": ", ", "))
		sb.WriteString(fmt.Sprintf("keys=%d", s.MaxKeys))
		n++
	}
	if s.MaxBytes > 0 {
		sb.WriteString(ext.Choose(n == 0, ": ", ", "))
		sb.WriteString(fmt.Sprintf("bytes=%s", humanize.Bytes(s.MaxBytes)))
		n++
	}
	return sb.String()
}

// evict is called by LRU when an element is evict. We updated our
// bookkeeping to free its' associated capacity.
func (s *Store) evict(_ string, v []byte) {
	s.Lock()
	s.nbytes -= uint64(len(v))
	s.Unlock()
}

// alloc determines if there is sufficient capacity to store the specified
// number of bytes. If so, it increments the consumed capacity and returns true.
// If not, it returns false.
func (s *Store) alloc(n uint64) (res bool) {
	s.Lock()
	res = s.MaxBytes <= 0 || (s.nbytes+n) <= s.MaxBytes
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
	if s.MaxBytes <= 0 {
		return 0, 0 // capacity is not limited; return zeros
	}
	s.Lock()
	n := s.nbytes
	s.Unlock()
	return s.MaxBytes, n
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
		return nil, fmt.Errorf("Not found: %s: %w", key, kvs.ErrNotFound)
	}
	return val, nil
}

func (s *Store) Set(cxt context.Context, key string, val []byte, opts ...kvs.WriteOption) error {
	// If we have a memory consumption limit, make a limitied number of attempts
	// to free enough capacity for the insertion. If we cannot do so, the
	// insertion fails.
	if s.MaxBytes > 0 {
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
	var (
		sum int64
		err error
	)
	val, ok := s.cache.Get(key)
	if !ok {
		val = []byte("0") // initialize to implied zero value
	}
	for i := 0; i < maxAttempts; i++ {
		sum, err = strconv.ParseInt(string(val), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("Incremented value must be an integer: %s: %w", key, err)
		}
		sum += inc // increment by the specified amount
		_, err = s.cache.Swap(key, []byte(strconv.FormatInt(sum, 10)), func(prev, curr []byte) error {
			if bytes.Compare(prev, val) != 0 {
				val = prev // update the current value for the next attempt
				return errCompareFailed
			}
			return nil
		})
		if err == errCompareFailed {
			continue
		}
		break
	}
	return sum, nil
}

func (s *Store) Delete(cxt context.Context, key string, opts ...kvs.WriteOption) error {
	// If the element exists in the LRU, the eviction handler is called to deal
	// with freeing capacity. If the element does not exist in the LRU, this
	// operation has no effect.
	s.cache.Remove(key)
	return nil
}
