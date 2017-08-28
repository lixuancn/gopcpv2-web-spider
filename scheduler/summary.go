package scheduler

import (
	"gopcpv2-web-spider/module"
	"sort"
	"encoding/json"
)

type SchedSummary interface {
	Struct() SummaryStruct
	String() string
}

type SummaryStruct struct{
	RequestArgs RequestArgs `json:"request_json"`
	DataArgs DataArgs `json:"data_args"`
	ModuleArgs ModuleArgs `json:"module_args"`
	Status string `json:"status"`
	Downloaders []module.Downloader `json:"downloader"`
	Analyzers []module.Analyzer `json:"analyzer"`
	Pipelines []module.Pipeline `json:"pipeline"`
	ReqBufferPool BufferPoolSummaryStruct `json:"request_buffer_pool"`
	RespBufferPool BufferPoolSummaryStruct `json:"response_buffer_pool"`
	ItemBufferPool BufferPoolSummaryStruct `json:"item_buffer_pool"`
	BufferPool BufferPoolSummaryStruct `json:"error_buffer_pool"`
	NumURL uint64 `json:"url_number"`
}

func newSchedSummary(requestArgs RequestArgs, dataArgs DataArgs, moduleArgs ModuleArgs, sched *myScheduler)SchedSummary{
	if sched == nil{
		return nil
	}
	return &mySchedSummary{
		requestArgs: requestArgs,
		dataArgs: dataArgs,
		moduleArgs: moduleArgs,
		sched: sched,
	}
}

type mySchedSummary struct {
	requestArgs RequestArgs
	dataArgs DataArgs
	moduleArgs ModuleArgs
	maxDepth uint32
	sched *myScheduler
}