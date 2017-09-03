package monitor

import (
	sched "gopcpv2-web-spider/scheduler"
	"time"
	"errors"
	"context"
	"fmt"
	"runtime"
)

type Record func(level uint8, content string)

func Monitor(scheduler sched.Scheduler, checkInterval time.Duration, summarizeInterval time.Duration, maxIdleCount uint, autoStop bool, record Record) <-chan uint64{
	if scheduler == nil{
		panic(errors.New("调度器不可用"))
	}
	if checkInterval < time.Microsecond * 100 {
		checkInterval = time.Microsecond * 100
	}
	if summarizeInterval < time.Second{
		summarizeInterval = time.Second
	}
	if maxIdleCount < 10 {
		maxIdleCount = 10
	}
	stopNotifier, stopFunc := context.WithCancel(context.Background())
	// 接收和报告错误。
	reportError(scheduler, record, stopNotifier)
	// 记录摘要信息。
	recordSummary(scheduler, summarizeInterval, record, stopNotifier)
	// 检查计数通道
	checkCountChan := make(chan uint64, 2)
	// 检查空闲状态
	checkStatus(scheduler, checkInterval, maxIdleCount, autoStop,
		checkCountChan, record, stopFunc)
	return checkCountChan
}

var msgReachMaxIdleCount = "调度器已经长时间空闲 (about %s)." + " ，准备关闭"

// msgStopScheduler 代表停止调度器的消息模板。
var msgStopScheduler = "正在停止调度器...%s."

func checkStatus(scheduler sched.Scheduler, checkInterval time.Duration,
	maxIdleCount uint, autoStop bool, checkCountChan chan<- uint64,
	record Record, stopFunc context.CancelFunc) {
	go func() {
		var checkCount uint64
		defer func() {
			stopFunc()
			checkCountChan <- checkCount
		}()
		// 等待调度器开启。
		waitForSchedulerStart(scheduler)
		// 准备。
		var idleCount uint
		var firstIdleTime time.Time
		for {
			// 检查调度器的空闲状态。
			if scheduler.Idle() {
				idleCount++
				if idleCount == 1 {
					firstIdleTime = time.Now()
				}
				if idleCount >= maxIdleCount {
					record(0, fmt.Sprintf(msgReachMaxIdleCount, time.Since(firstIdleTime).String()))
					// 再次检查调度器的空闲状态，确保它已经可以被停止。
					if scheduler.Idle() {
						if autoStop {
							var result string
							if err := scheduler.Stop(); err == nil {
								result = "成功"
							} else {
								result = fmt.Sprintf("失败(%s)", err)
							}
							record(0, fmt.Sprintf(msgStopScheduler, result))
						}
						break
					} else {
						if idleCount > 0 {
							idleCount = 0
						}
					}
				}
			} else {
				if idleCount > 0 {
					idleCount = 0
				}
			}
			checkCount++
			time.Sleep(checkInterval)
		}
	}()
}

// summary 代表监控结果摘要的结构。
type summary struct {
	// NumGoroutine 代表Goroutine的数量。
	NumGoroutine int `json:"goroutine_number"`
	// SchedSummary 代表调度器的摘要信息。
	SchedSummary sched.SummaryStruct `json:"sched_summary"`
	// EscapedTime 代表从开始监控至今流逝的时间。
	EscapedTime string `json:"escaped_time"`
}


func recordSummary(scheduler sched.Scheduler, summarizeInterval time.Duration, record Record, stopNotifier context.Context) {
	go func() {
		waitForSchedulerStart(scheduler)
		var prevSchedSummaryStruct sched.SummaryStruct
		var prevNumGoroutine int
		var recordCount uint64 = 1
		//startTime := time.Now()
		for {
			select {
			case <-stopNotifier.Done():
				return
			default:
			}
			currNumGoroutine := runtime.NumGoroutine()
			currSchedSummaryStruct := scheduler.Summary().Struct()
			if currNumGoroutine != prevNumGoroutine || !currSchedSummaryStruct.Same(prevSchedSummaryStruct) {
				//summay := summary{
				//	NumGoroutine: runtime.NumGoroutine(),
				//	SchedSummary: currSchedSummaryStruct,
				//	EscapedTime: time.Since(startTime).String(),
				//}
				//b, err := json.MarshalIndent(summay, "", "    ")
				//if err != nil {
				//	fmt.Println(fmt.Sprintf("生成摘要时发生了错误: %s\n", err))
				//	continue
				//}
				//record(0, fmt.Sprintf("监控摘要[%d]:\n%s", recordCount, b))
				prevNumGoroutine = currNumGoroutine
				prevSchedSummaryStruct = currSchedSummaryStruct
				recordCount++
			}
			time.Sleep(summarizeInterval)
		}
	}()
}

func reportError(scheduler sched.Scheduler, record Record, stopNotifier context.Context){
	go func() {
		waitForSchedulerStart(scheduler)
		errorChan := scheduler.ErrorChan()
		for{
			select{
			case <- stopNotifier.Done():
				return
			default:
			}
			err, ok := <-errorChan
			if ok{
				record(2, fmt.Sprintf("获取一个错误从错误通道中：%s", err))
			}
			time.Sleep(time.Microsecond)
		}
	}()
}

func waitForSchedulerStart(scheduler sched.Scheduler) {
	for scheduler.Status() != sched.SCHED_STATUS_STARTED {
		time.Sleep(time.Microsecond)
	}
}