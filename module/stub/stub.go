package stub

import (
	"gopcpv2-web-spider/module"
	"fmt"
	"sync/atomic"
	"gopcpv2-web-spider/errors"
)

type myModule struct {
	mid module.MID
	addr string
	score uint64
	scoreCalculator module.CalculateScore
	calledCount uint64
	acceptedCount uint64
	completedCount uint64
	handlingNumber uint64
}

func NewModuleInternal(mid module.MID, scoreCalculator module.CalculateScore)(ModuleInternal, error){
	parts, err := module.SplitMID(mid)
	if err != nil{
		return nil, errors.NewIllegalParameterError(fmt.Sprintf("无效的ID%q: %s", mid, err))
	}
	return &myModule{
		mid: mid,
		addr: parts[2],
		scoreCalculator: scoreCalculator,
	}, nil
}

func (m *myModule) ID() module.MID {
	return m.mid
}

func (m *myModule) Addr() string {
	return m.addr
}

func (m *myModule) Score() uint64 {
	return atomic.LoadUint64(&m.score)
}

func (m *myModule) SetScore(score uint64) {
	atomic.StoreUint64(&m.score, score)
}

func (m *myModule) ScoreCalculator() module.CalculateScore {
	return m.scoreCalculator
}

func (m *myModule) CalledCount() uint64 {
	return atomic.LoadUint64(&m.calledCount)
}

func (m *myModule) AcceptedCount() uint64 {
	return atomic.LoadUint64(&m.acceptedCount)
}

func (m *myModule) CompletedCount() uint64 {
	return atomic.LoadUint64(&m.completedCount)
}

func (m *myModule) HandlingNumber() uint64 {
	return atomic.LoadUint64(&m.handlingNumber)
}

func (m *myModule) Counts() module.Counts {
	return module.Counts{
		CalledCount:    atomic.LoadUint64(&m.calledCount),
		AcceptedCount:  atomic.LoadUint64(&m.acceptedCount),
		CompletedCount: atomic.LoadUint64(&m.completedCount),
		HandlingNumber: atomic.LoadUint64(&m.handlingNumber),
	}
}

func (m *myModule) Summary() module.SummaryStruct {
	counts := m.Counts()
	return module.SummaryStruct{
		ID:        m.ID(),
		Called:    counts.CalledCount,
		Accepted:  counts.AcceptedCount,
		Completed: counts.CompletedCount,
		Handling:  counts.HandlingNumber,
		Extra:     nil,
	}
}


func (m *myModule)IncrCalledCount(){
	atomic.AddUint64(&m.calledCount, 1)
}

func (m *myModule)IncrAcceptedCount(){
	atomic.AddUint64(&m.acceptedCount, 1)
}

func (m *myModule)IncrCompletedCount(){
	atomic.AddUint64(&m.completedCount, 1)
}

func (m *myModule)IncrHandlingNumber(){
	atomic.AddUint64(&m.handlingNumber, 1)
}

func (m *myModule)DecrHandlingNumber(){
	atomic.AddUint64(&m.handlingNumber, ^uint64(0))
}

func (m *myModule)Clear(){
	atomic.StoreUint64(&m.calledCount, 0)
	atomic.StoreUint64(&m.acceptedCount, 0)
	atomic.StoreUint64(&m.completedCount, 0)
	atomic.StoreUint64(&m.handlingNumber, 0)
}