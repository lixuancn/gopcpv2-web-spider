package module

import (
	"math"
	"sync"
	"sync/atomic"
)

type SNGenertor interface {
	//最小序列号
	Start() uint64
	Max() uint64
	//下一个序列号
	Next()uint64
	//用户获取循环次数，即从前两名方法给定范围的循环次数
	CycleCount() uint64
	//获取一个序列号并准备下一个序列号
	Get()uint64
}

type mySNGenertor struct {
	start uint64
	max uint64
	next uint64
	cycleCount uint64
	lock sync.RWMutex
}

func NewSNGenertor(start, max uint64)SNGenertor{
	if max == 0{
		max = math.MaxUint64
	}
	return &mySNGenertor{
		start: start,
		max: max,
		next: start,
	}
}

func (gen *mySNGenertor)Start()uint64{
	return gen.start
}

func (gen *mySNGenertor)Max()uint64{
	return gen.max
}

func (gen *mySNGenertor)Next()uint64{
	//gen.lock.RLock()
	//defer gen.lock.RUnlock()
	//return gen.next
	return atomic.LoadUint64(&gen.next)
}

func (gen *mySNGenertor)CycleCount()uint64{
	//gen.lock.RLock()
	//defer gen.lock.RUnlock()
	//return gen.cycleCount
	return atomic.LoadUint64(&gen.cycleCount)
}

func (gen *mySNGenertor)Get()uint64{
	gen.lock.Lock()
	defer gen.lock.Unlock()
	id := gen.next
	if id == gen.max{
		gen.next = gen.start
		gen.cycleCount++
	}else{
		gen.next++
	}
	return id
}