package scheduler

import "gopcpv2-web-spider/module"

type Args interface {
	Check()error
}

type RequestArgs struct{
	//主机列表
	AcceptedDomains []string `json:"accepted_primary_domains"`
	//最大深度
	MaxDepth uint32 `json:"max_depth"`
}

func (args *RequestArgs)Check()error{
	if  args.AcceptedDomains == nil || len(args.AcceptedDomains) <= 0{
		return genError("接受的域名列表不能是nil")
	}
	return nil
}

// Same 用于判断两个请求相关的参数容器是否相同。
func (args *RequestArgs) Same(another *RequestArgs) bool {
	if another == nil {
		return false
	}
	if another.MaxDepth != args.MaxDepth {
		return false
	}
	if len(another.AcceptedDomains) != len(args.AcceptedDomains) {
		return false
	}
	if len(another.AcceptedDomains) > 0 {
		for i, domain := range another.AcceptedDomains {
			if domain != args.AcceptedDomains[i] {
				return false
			}
		}
	}
	return true
}

type DataArgs struct{
	ReqBufferCap uint32 `json:"req_buffer_cap"`
	ReqMaxBufferNumber uint32 `json:"req_max_buffe_number"`
	RespBufferCap uint32 `json:"resp_buffer_cap"`
	RespMaxBufferNumber uint32 `json:"resp_max_buffer_number"`
	ItemBufferCap uint32 `json:"item_buffer_cap"`
	ItemMaxBufferNumber uint32 `json:"item_max_buffer_number"`
	ErrorBufferCap uint32 `json:"error_buffer_cap"`
	ErrorMaxBufferNumber uint32 `json:"error_max_buffer_number"`
}

func (args *DataArgs)Check()error{
	if args.ReqBufferCap == 0 {
		return genError("zero request buffer capacity")
	}
	if args.ReqMaxBufferNumber == 0 {
		return genError("zero max request buffer number")
	}
	if args.RespBufferCap == 0 {
		return genError("zero response buffer capacity")
	}
	if args.RespMaxBufferNumber == 0 {
		return genError("zero max response buffer number")
	}
	if args.ItemBufferCap == 0 {
		return genError("zero item buffer capacity")
	}
	if args.ItemMaxBufferNumber == 0 {
		return genError("zero max item buffer number")
	}
	if args.ErrorBufferCap == 0 {
		return genError("zero error buffer capacity")
	}
	if args.ErrorMaxBufferNumber == 0 {
		return genError("zero max error buffer number")
	}
	return nil
}

type ModuleArgs struct {
	Downloaders []module.Downloader
	Analyzers []module.Analyzer
	Pipelines []module.Pipeline
}

func (args *ModuleArgs)Check()error{
	if len(args.Downloaders) == 0 {
		return genError("empty downloader list")
	}
	if len(args.Analyzers) == 0 {
		return genError("empty analyzer list")
	}
	if len(args.Pipelines) == 0 {
		return genError("empty pipeline list")
	}
	return nil
}


type ModuleArgsSummary struct {
	DownloaderListSize int `json:"downloader_list_size"`
	AnalyzerListSize   int `json:"analyzer_list_size"`
	PipelineListSize   int `json:"pipeline_list_size"`
}

func (args *ModuleArgs) Summary() ModuleArgsSummary {
	return ModuleArgsSummary{
		DownloaderListSize: len(args.Downloaders),
		AnalyzerListSize:   len(args.Analyzers),
		PipelineListSize:   len(args.Pipelines),
	}
}