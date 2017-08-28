package scheduler

import "net/http"

type Scheduler interface {
	Init(requestArgs RequestArgs, dataArgs DataArgs, moduleArgs ModuleArgs)
	Start(firstHttpReq *http.Request)error
	Stop()error
	Status()Status
	ErrorChan()<-chan error
	Idel()bool
	Summary()SchedSummary
}