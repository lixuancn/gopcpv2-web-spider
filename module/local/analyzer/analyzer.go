package analyzer

import (
	"gopcpv2-web-spider/module/stub"
	"gopcpv2-web-spider/module"
	"fmt"
	"log"
	"gopcpv2-web-spider/toolkit/reader"
)

type myAnalyzer struct {
	stub.ModuleInternal
	respParsers []module.ParseResponse
}

func New(mid module.MID, respParsers []module.ParseResponse, scoreCalculator module.CalculateScore)(module.Analyzer, error){
	moduleBase, err := stub.NewModuleInternal(mid, scoreCalculator)
	if err != nil{
		return nil, err
	}
	if respParsers == nil{
		return nil, genParameterError("解析响应函数列表是nil")
	}
	if len(respParsers) == 0{
		return nil, genParameterError("解析响应函数列表为空")
	}
	var innerParsers []module.ParseResponse
	for i, parser := range respParsers{
		if parser == nil{
			return nil, genParameterError(fmt.Sprintf("解析响应函数列表[%d]是nil", i))
		}
		innerParsers = append(innerParsers, parser)
	}
	return &myAnalyzer{
		ModuleInternal: moduleBase,
		respParsers: innerParsers,
	}, nil
}


func(analyzer *myAnalyzer)RespParsers()[]module.ParseResponse {
	parsers := make([]module.ParseResponse, len(analyzer.respParsers))
	copy(parsers, analyzer.respParsers)
	return parsers
}

func(analyzer *myAnalyzer)Analyze(resp *module.Response)(dataList []module.Data, errorList []error){
	analyzer.ModuleInternal.IncrHandlingNumber()
	defer analyzer.ModuleInternal.DecrHandlingNumber()
	analyzer.ModuleInternal.IncrCalledCount()
	if resp == nil{
		errorList = append(errorList, genParameterError("响应是nil"))
		return
	}
	httpResp := resp.HTTPResp()
	if httpResp == nil{
		errorList = append(errorList, genParameterError("HTTP响应是nil"))
		return
	}
	httpReq := httpResp.Request
	if httpReq == nil{
		errorList = append(errorList, genParameterError("HTTP请求是nil"))
		return
	}
	var reqUrl = httpReq.URL
	if reqUrl == nil{
		errorList = append(errorList, genParameterError("HTTP请求的URL是nil"))
		return
	}
	analyzer.ModuleInternal.IncrAcceptedCount()
	log.Printf("解析响应, URL: %s, depth: %d...", reqUrl, resp.Depth())
	if httpResp.Body != nil{
		defer httpResp.Body.Close()
	}
	multipleReader, err := reader.NewMultipleReader(httpResp.Body)
	if err != nil{
		errorList = append(errorList, genParameterError(err.Error()))
		return
	}
	dataList = []module.Data{}
	for _, respParser := range analyzer.respParsers{
		httpResp.Body = multipleReader.Reader()
		pDataList, pErrorList := respParser(httpResp, resp.Depth())
		if pDataList != nil{
			for _, pData := range pDataList{
				if pData == nil{
					continue
				}
				dataList = appendDataList(dataList, pData, resp.Depth())
			}
		}
		if pErrorList != nil{
			for _, pError := range pErrorList{
				if pError == nil{
					continue
				}
				errorList = append(errorList, pError)
			}
		}
	}
	if len(errorList) == 0{
		analyzer.ModuleInternal.IncrCompletedCount()
	}
	return
}

func appendDataList(dataList []module.Data, data module.Data, respDepth uint32) []module.Data {
	if data == nil {
		return dataList
	}
	req, ok := data.(*module.Request)
	if !ok {
		return append(dataList, data)
	}
	newDepth := respDepth + 1
	if req.Depth() != newDepth {
		req = module.NewRequest(req.HTTPReq(), newDepth)
	}
	return append(dataList, req)
}