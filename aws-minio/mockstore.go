package main

import (
	"io"
	"sync"
)

type mockStore struct {
	sync.Mutex
	fileCounter int
}

func NewMockStore() {

}

func (ms *mockStore) Put(key string, body io.Reader) (string, error) {
	ms.Lock()
	ms.fileCounter++
	ms.Unlock()
	return "", nil
}
