package stub

import (
	"gopcpv2-web-spider/module"
)

type ModuleInternal interface {
	module.Module
	IncrCalledCount()
	IncrAcceptedCount()
	IncrCompletedCount()
	IncrHandlingNumber()
	DecrHandlingNumber()
	Clear()
}