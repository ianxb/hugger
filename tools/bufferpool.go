package tools

import (
	"sync"
	"fmt"
	"errors"
	"sync/atomic"
)


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

type myPool struct {
	bufferCap       uint32       //the capacity of buffer
	maxBufferNumber uint32       //the number of buffer
	bufferNumber    uint32       //current number of buffer
	total           uint64       //the data number
	bufChan         chan Buffer  //the channel for buffer
	closed          uint32        //status check variable: atomic operation
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

func (this *myPool) Close() bool{
	if !atomic.CompareAndSwapUint32(&this.closed, 0,1){
		return false
	}
	this.rwlock.Lock()
	defer this.rwlock.Unlock()
	close(this.bufChan)
	for buf := range this.bufChan{
		buf.Close()
	}
	return true
}

