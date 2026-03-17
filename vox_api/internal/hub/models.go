package hub

import "sync"

type Channel interface {
	Close()
}

type StringChan struct {
	mu       *sync.Mutex
	Ch       chan string
	isClosed bool
}

type ErrorChan struct {
	mu       *sync.Mutex
	Ch       chan error
	isClosed bool
}

func (sc *StringChan) Close() {
	sc.mu.Lock()
	if !sc.isClosed {
		close(sc.Ch)
		sc.isClosed = true
	}
	sc.mu.Unlock()
}

func (ec *ErrorChan) Close() {
	ec.mu.Lock()
	if !ec.isClosed {
		close(ec.Ch)
		ec.isClosed = true
	}
	ec.mu.Unlock()
}

func NewStringChanBuf(size int) *StringChan {
	return &StringChan{
		mu: &sync.Mutex{},
		Ch: make(chan string, size),
	}
}

func NewErrorChanBuf(size int) *ErrorChan {
	return &ErrorChan{
		mu: &sync.Mutex{},
		Ch: make(chan error, size),
	}
}
