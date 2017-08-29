package buffer

import (
	"sync"
	"gopcpv2-web-spider/errors"
	"fmt"
	"sync/atomic"
)

type Buffer interface {
	Cap() uint32
	Len() uint32
	//非阻塞
	Put(interface{})(bool, error)
	//非阻塞
	Get()(interface{}, error)
	Close()bool
	Closed()bool
}

type myBuffer struct {
	ch chan interface{}
	closed uint32
	closeingLock sync.RWMutex
}

func NewBuffer(size uint32)(Buffer, error){
	if size == 0{
		return nil, errors.NewIllegalParameterError(fmt.Sprintf("缓冲器尺寸无效: %d", size))
	}
	return &myBuffer{ch: make(chan interface{}, size)}, nil
}

func (buf *myBuffer)Cap()uint32{
	return uint32(cap(buf.ch))
}

func (buf *myBuffer)Len()uint32{
	return uint32(len(buf.ch))
}

func (buf *myBuffer)Close()bool{
	if atomic.CompareAndSwapUint32(&buf.closed, 0, 1){
		buf.closeingLock.Lock()
		close(buf.ch)
		defer buf.closeingLock.Unlock()
		return true
	}
	return false
}

func (buf *myBuffer)Closed()bool{
	if atomic.LoadUint32(&buf.closed) == 0{
		return false
	}
	return true
}

func (buf *myBuffer)Put(datum interface{})(bool, error){
	//这个锁只判断关闭状态，而不是用来保证写入的
	buf.closeingLock.RLock()
	defer buf.closeingLock.RUnlock()
	if buf.Closed() {
		return false, ErrClosedBuffer
	}
	var ok bool
	select{
	case buf.ch <- datum:
		ok = true
	default:
		ok = false
	}
	return ok, nil
}

func (buf *myBuffer)Get()(interface{}, error){
	select {
	case datum, ok := <- buf.ch:
		if !ok{
			return nil, ErrClosedBuffer
		}
		return datum, nil
	default:
		return nil, nil
	}
}