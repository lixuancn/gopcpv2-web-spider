package buffer

import (
	"sync"
	"gopcpv2-web-spider/errors"
	"sync/atomic"
)

type Pool interface {
	//统一的缓冲器的容量
	BufferCap() uint32
	//缓冲器的最大数量
	MaxBufferNumber() uint32
	//缓冲器的数量
	BufferNumber() uint32
	//缓冲池中的数据总数
	Total() uint64
	//入，阻塞的
	Put(interface{}) error
	//取，阻塞的
	Get()(interface{}, error)
	Close() bool
	Closed() bool
}

type myPool struct {
	bufferCap uint32
	maxBufferNumber uint32
	bufferNumber uint32
	total uint64
	bufCh chan Buffer
	closed uint32
	rwlock sync.RWMutex
}

func NewPool(bufferCap, maxBufferNumber uint32)(Pool, error){
	if bufferCap == 0{
		return nil, errors.NewIllegalParameterError("缓冲池中缓冲器统一容量不可以为0")
	}
	if maxBufferNumber == 0{
		return nil, errors.NewIllegalParameterError("缓冲池中缓冲器最大容量不可以为0")
	}
	bufCh := make(chan Buffer, maxBufferNumber)
	buf, err := NewBuffer(bufferCap)
	if err != nil{
		return nil, err
	}
	bufCh <- buf
	return &myPool{
		bufferCap:       bufferCap,
		maxBufferNumber: maxBufferNumber,
		bufferNumber:    1,
		bufCh:           bufCh,
	}, nil
}


func (pool *myPool) BufferCap() uint32 {
	return pool.bufferCap
}

func (pool *myPool) MaxBufferNumber() uint32 {
	return pool.maxBufferNumber
}

func (pool *myPool) BufferNumber() uint32 {
	return atomic.LoadUint32(&pool.bufferNumber)
}

func (pool *myPool) Total() uint64 {
	return atomic.LoadUint64(&pool.total)
}

func (pool *myPool)Close()bool{
	if !atomic.CompareAndSwapUint32(&pool.closed, 0, 1){
		return false
	}
	pool.rwlock.Lock()
	defer pool.rwlock.Unlock()
	close(pool.bufCh)
	for buf := range pool.bufCh{
		buf.Close()
	}
	return true
}

func (pool *myPool)Closed()bool{
	if atomic.LoadUint32(&pool.closed) == 1{
		return true
	}
	return false
}

func (pool *myPool)Put(datum interface{})(error){
	if pool.Closed(){
		return ErrClosedBufferPool
	}
	var count uint32
	maxCount := pool.BufferNumber() * 5
	var err error
	for buf := range pool.bufCh{
		ok, err := pool.putData(buf, datum, &count, maxCount)
		if ok || err != nil{
			break
		}
	}
	return err
}

func (pool *myPool)putData(buf Buffer, datum interface{}, count *uint32, maxCount uint32)(bool, error){
	if pool.Closed(){
		return false, ErrClosedBufferPool
	}
	var err error
	var ok bool
	defer func(){
		pool.rwlock.RLock()
		defer pool.rwlock.RUnlock()
		if pool.Closed(){
			atomic.AddUint32(&pool.bufferNumber, ^uint32(0))
			err = ErrClosedBufferPool
		}else{
			pool.bufCh <- buf
		}
	}()
	ok, err = buf.Put(datum)
	if ok {
		atomic.AddUint64(&pool.total, 1)
		return true, nil
	}
	if err != nil{
		return false, err
	}
	(*count)++
	if *count >= maxCount && pool.MaxBufferNumber() > pool.BufferNumber(){
		pool.rwlock.Lock()
		if pool.MaxBufferNumber() > pool.BufferNumber(){
			if pool.Closed(){
				pool.rwlock.Unlock()
				return false, err
			}
			newBuf, err := NewBuffer(pool.BufferCap())
			if err != nil{
				return false, err
			}
			newBuf.Put(datum)
			pool.bufCh <- newBuf
			atomic.AddUint32(&pool.bufferNumber, 1)
			atomic.AddUint64(&pool.total, 1)
			ok = true
		}
		pool.rwlock.Unlock()
		*count = 0
	}
	return ok, err
}

func (pool *myPool)Get()(interface{}, error){
	if pool.Closed(){
		return nil, ErrClosedBufferPool
	}
	var datum interface{}
	var err error
	var count uint32
	maxCount := pool.BufferNumber() * 10
	for buf := range pool.bufCh{
		datum, err = pool.getData(buf, &count, maxCount)
		if datum != nil || err != nil{
			break
		}
	}
	return datum, err
}

func (pool *myPool)getData(buf Buffer, count *uint32, maxCount uint32)(datum interface{}, err error){
	if pool.Closed(){
		err = ErrClosedBufferPool
		return
	}
	defer func(){
		if *count > maxCount && buf.Len() == 0 && pool.BufferNumber() > 1{
			buf.Close()
			atomic.AddUint32(&pool.bufferNumber, ^uint32(0))
			*count = 0
			return
		}
		pool.rwlock.RLock()
		defer pool.rwlock.RUnlock()
		if pool.Closed(){
			atomic.AddUint32(&pool.bufferNumber, ^uint32(0))
			err = ErrClosedBufferPool
		}else{
			pool.bufCh <- buf
		}
	}()
	datum, err = buf.Get()
	if err != nil{
		return
	}
	if datum != nil{
		atomic.AddUint64(&pool.total, ^uint64(0))
		return
	}
	(*count)++
	return
}