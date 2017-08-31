package downloader

import (
	"gopcpv2-web-spider/module"
	"net/http"
	"gopcpv2-web-spider/module/stub"
	"log"
)

type myDownloader struct{
	stub.ModuleInternal
	httpClient http.Client
}

func New(mid module.MID, client *http.Client, scoreCalculator module.CalculateScore)(module.Downloader, error){
	moduleBase, err := stub.NewModuleInternal(mid, scoreCalculator)
	if err != nil{
		return nil, err
	}
	if client == nil{
		return nil, genParameterError("HttpClient是nil")
	}
	return &myDownloader{
		ModuleInternal: moduleBase,
		httpClient: *client,
	}, nil
}

func (downloader *myDownloader)Download(req *module.Request)(*module.Response, error){
	downloader.ModuleInternal.IncrHandlingNumber()
	defer downloader.ModuleInternal.DecrHandlingNumber()
	downloader.ModuleInternal.IncrCalledCount()
	if req == nil{
		return nil, genParameterError("请求是nil")
	}
	httpReq := req.HTTPReq()
	if httpReq == nil{
		return nil, genParameterError("HTTP请求是nil")
	}
	downloader.ModuleInternal.IncrAcceptedCount()
	log.Printf("执行请求：url: %s, depth: %d...\n", httpReq.URL, req.Depth())
	httpResp, err := downloader.httpClient.Do(httpReq)
	if err != nil{
		return nil, err
	}
	downloader.ModuleInternal.IncrCompletedCount()
	return module.NewResponse(httpResp, req.Depth()), nil
}