package tools

import (
	"sync"
	//"github.com/chaozh/MIT-6.824-2017/src/kvraft"
	"fmt"
	//"errors"
	//"golang.org/x/tools/go/gcimporter15/testdata"
	"errors"
	"sync/atomic"
)

//import "github.com/chaozh/MIT-6.824-2017/src/kvraft"

type Pool interface {
	BufferCap() uint32
	MaxBufferNumber() uint32
	BufferNumber() uint32
	Total() uint64
	Put(data interface{}) error         //put data into pool, if the pool is closed return error
	Get() (data interface{}, err error) //get data from pool
	Close() bool                        //closed a pool
	Closed() bool                       //find whether a pool is closed
}

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

type myPool struct {
	bufferCap       uint32       //the capacity of buffer
	maxBufferNumber uint32       //the number of buffer
	bufferNumber    uint32       //current number of buffer
	total           uint64       //the data number
	bufChan         chan Buffer  //the channel for buffer
	closed          int32        //status check variable: atomic operation
	rwlock          sync.RWMutex //lock for current operation
}

func NewPool(bufferCap uint32, maxBufferNumber uint32) (pool Pool, err error) {
	if bufferCap == 0 {
		err := fmt.Sprintf("the capacity of pool is: %s", bufferCap)
		return nil, error.NewIllegalParError(err)
	}
	if maxBufferNumber == 0 {
		err := fmt.Sprintf("the number of buffer is: %s", maxBufferNumber)
		return nil, error.NewIllegalParError(err)
	}
	bufChan := make(chan Buffer, maxBufferNumber)
	buffer, _ := NewBuffer(bufferCap)
	bufChan <- buffer
	return &myPool{bufferCap: bufferCap, maxBufferNumber: maxBufferNumber, bufferNumber: 1, bufChan: bufChan}, nil
}

func (this *myPool) BufferCap() uint32 {
	return this.bufferCap
}
func (this *myPool) MaxBufferNumber() uint32 {
	return this.maxBufferNumber
}
func (this *myPool) BufferNumber() uint32 {
	return atomic.LoadUint32(&this.bufferNumber)
}
func (this *myPool) Total() uint64 {
	return atomic.LoadUint64(&this.total)
}
func (this *myPool) Closed() bool {
	if atomic.LoadInt32(&this.closed) == 0 {
		return true
	}
	return false
}

var ErrclosedBufferPool = errors.New("closed buffer pool error")

func (this *myPool) Put(data interface{}) (err error) {
	if this.Closed() {
		return ErrClosedBuffer
	}
	var maxCount = this.BufferNumber() * 5
	var count uint32
	var ok bool
	for buf := range this.bufChan {
		ok, err = this.putData(buf, data, &count, maxCount)
		if ok || err != nil {
			break
		}
	}
	return
}

func (this *myPool) putData(buf Buffer, data interface{}, count *uint32, maxCount uint32) (ok bool, err error) {
	if this.Closed() {
		return false, ErrclosedBufferPool
	}
	defer func() {
		this.rwlock.Lock()
		if this.Closed() {
			atomic.AddUint32(&this.bufferNumber, ^uint32(0))
			err = ErrClosedBuffer
		} else {
			this.bufChan <- buf
		}
		this.rwlock.Unlock()
	}()

	ok, err = buf.Put(data)
	if ok {
		atomic.AddUint64(&this.total, 1)
		return
	}
	if err != nil {
		return
	}
	(*count)++
	if *count >= maxCount && this.BufferNumber() < this.MaxBufferNumber() {
		this.rwlock.Lock()
		if this.BufferNumber() < this.MaxBufferNumber() {
			if this.Closed() {
				this.rwlock.Unlock()
				return
			}
			newBuf, _ := NewBuffer(this.bufferCap)
			newBuf.Put(data)
			this.bufChan <- newBuf
			atomic.AddUint32(&this.bufferNumber, 1)
			atomic.AddUint64(&this.total, 1)
			ok = true
		}
		this.rwlock.Unlock()
		*count = 0
	}
	return
}

func (this *myPool) Get() (data interface{}, err error) {
	if this.Closed() {
		return nil, ErrclosedBufferPool
	}
	var count uint32
	maxCount := this.BufferNumber() * 10
	for buf := range this.bufChan {
		data, err = this.getData(buf, &count, maxCount)
		if data != nil || err != nil {
			break
		}
	}
	return
}

func (this *myPool) getData(buf Buffer, count *uint32, maxCount uint32) (data interface{}, err error) {
	if this.Closed() {
		return nil, ErrclosedBufferPool
	}
	defer func() {
		if *count >= maxCount && this.BufferNumber() > 1 {
			buf.Close()
			atomic.AddUint32(&this.bufferNumber, ^uint32(0))
			return
		}
		this.rwlock.Lock()
		if this.Closed() {
			atomic.AddUint32(&this.bufferNumber, uint32(0))
			err = ErrclosedBufferPool
		} else {
			this.bufChan <- buf
		}
		this.rwlock.Unlock()
	}()
	data, err = buf.Get()
	if data != nil {
		atomic.AddUint64(&this.total, ^uint64(0))
		return
	}
	if err != nil {
		return
	}
	(*count)++
	return
}
