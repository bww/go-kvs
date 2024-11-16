package memory

import "github.com/bww/go-kvs/v1"

type keysIter struct {
	data chan string
}

func (r keysIter) Next() (string, error) {
	v, ok := <-r.data
	if !ok {
		return "", kvs.ErrClosed
	} else {
		return v, nil
	}
}
