package pipeline

import (
	"gopcpv2-web-spider/module"
	"gopcpv2-web-spider/module/stub"
	"fmt"
	"log"
)

type myPipeline struct{
	stub.ModuleInternal
	itemProcessors []module.ProcessItem
	failFast bool
}

func New(mid module.MID, itemProcessors []module.ProcessItem, scoreCalculator module.CalculateScore)(module.Pipeline, error){
	moduleBase, err := stub.NewModuleInternal(mid, scoreCalculator)
	if err != nil{
		return nil, err
	}
	if itemProcessors == nil{
		return nil, genParameterError("条目处理函数列表为空")
	}
	var innerProcessors []module.ProcessItem
	for i, pipeline := range itemProcessors{
		if pipeline == nil{
			return nil, genParameterError(fmt.Sprintf("条目处理函数为nil，索引为%d", i))
		}
		innerProcessors = append(innerProcessors, pipeline)
	}
	return &myPipeline{
		ModuleInternal: moduleBase,
		itemProcessors: innerProcessors,
		failFast: false,
	}, nil
}

//返回当前条目处理管道使用的条目处理函数的列表
func (pipeline *myPipeline)ItemProcessors() []module.ProcessItem{
	processItemList := make([]module.ProcessItem, len(pipeline.itemProcessors))
	copy(processItemList, pipeline.itemProcessors)
	return processItemList
}

//发送条目，条目被发送后会依次经过若干的条目处理函数的处理
func (pipeline *myPipeline)Send(item module.Item)[]error{
	pipeline.ModuleInternal.IncrHandlingNumber()
	defer pipeline.ModuleInternal.DecrHandlingNumber()
	pipeline.ModuleInternal.IncrCalledCount()
	var errList []error
	if item == nil{
		errList = append(errList, genParameterError("条目为空"))
		return errList
	}
	pipeline.ModuleInternal.IncrAcceptedCount()
	log.Printf("条目处理开始，%+v...\n", item)
	var currentItem = item
	for _, processor := range pipeline.itemProcessors{
		processedItem, err := processor(currentItem)
		if err != nil{
			errList = append(errList, err)
			if pipeline.failFast{
				break
			}
		}
		if processedItem != nil{
			currentItem = processedItem
		}
	}
	if len(errList) == 0{
		pipeline.IncrCompletedCount()
	}
	return errList
}

//当前条目处理管道是否是快速失败的，快速失败是只要在某条目被处理的某一个步骤上出错，则该条目后续处理都会忽略
func (pipeline *myPipeline)FailFast()bool{
	return pipeline.failFast
}

func (pipeline *myPipeline)SetFailFast(failFast bool){
	pipeline.failFast = failFast
}

//代表条目处理管道实额外信息的摘要类型。
type extraSummaryStruct struct {
	FailFast        bool `json:"fail_fast"`
	ProcessorNumber int  `json:"processor_number"`
}

func (pipeline *myPipeline) Summary()module.SummaryStruct{
	summary := pipeline.ModuleInternal.Summary()
	summary.Extra = extraSummaryStruct{
		FailFast:        pipeline.failFast,
		ProcessorNumber: len(pipeline.itemProcessors),
	}
	return summary
}