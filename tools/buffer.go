package tools

import (
	"sync/atomic"
	"sync"
	"fmt"
	"errors"
)

type Buffer interface {
	Cap() uint32
	Len() uint32
	Put(data interface{}) (bool, error)
	Get() (interface{}, error)
	Close() bool
	Closed() bool
}

type myBuffer struct {
	ch          chan interface{}
	closed      int32        // 0 stand for closed, 1 stand for no closed
	closingLock sync.RWMutex //RW lock
}

func NewBuffer(size uint32) (Buffer, error) {
	if size == 0 {
		errMsg := fmt.Sprintf("illegal size for buffer: %d", size)
		return nil, error.NewIllegalParError(errMsg)
	}
	return &myBuffer{ch: make(chan interface{}, size)}, nil
}

func (this *myBuffer) Cap() uint32 {
	return uint32(cap(this.ch))
}
func (this *myBuffer) Len() uint32 {
	return uint32(len(this.ch))
}

var ErrClosedBuffer = errors.New("closed buffer")

func (this *myBuffer) Put(data interface{}) (ok bool, err error) {
	this.closingLock.RLock()
	defer this.closingLock.RUnlock()
	if this.Closed() {
		return false, ErrClosedBuffer
	}
	select {
	case this.ch <- data:
		ok = true
	default:
		ok = false
	}
	return
}

func (this *myBuffer) Get() (interface{}, error) {
	select {
	case data, ok := <-this.ch:
		if !ok {
			return nil, ErrClosedBuffer
		}
		return data, nil
	default:
		return nil, nil
	}
}

func (this *myBuffer) Close() bool {
	if atomic.CompareAndSwapInt32(&this.closed, 1, 0) { //need atomic operation!!!
		this.closingLock.Lock()
		close(this.ch)
		this.closingLock.Unlock()
		return true
	}
	return false
}

func (this *myBuffer) Closed() bool {
	if atomic.LoadInt32(&this.closed) == 0 {
		return true
	}
	return false
}

