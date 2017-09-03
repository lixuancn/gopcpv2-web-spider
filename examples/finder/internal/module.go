package internal

import (
	"gopcpv2-web-spider/module"
	"gopcpv2-web-spider/module/local/downloader"
	"gopcpv2-web-spider/module/local/analyzer"
	"gopcpv2-web-spider/module/local/pipeline"
)

var snGen = module.NewSNGenertor(1, 0)

func GetDownloaders(number uint8)([]module.Downloader, error){
	downloaders := []module.Downloader{}
	if number == 0{
		return downloaders, nil
	}
	for i:=uint8(0); i<number; i++{
		mid, err := module.GenMID(module.TYPE_DOWNLOADER, snGen.Get(), nil)
		if err != nil{
			return downloaders, nil
		}
		d, err := downloader.New(mid, genHttpClient(), module.CalculateScoreSimple)
		if err != nil{
			return downloaders, nil
		}
		downloaders = append(downloaders, d)
	}
	return downloaders, nil
}

func GetAnalyzer(number uint8)([]module.Analyzer, error){
	analyzerList := []module.Analyzer{}
	if number  == 0{
		return analyzerList, nil
	}
	for i:=uint8(0); i<number; i++{
		mid, err := module.GenMID(module.TYPE_ANALYZER, snGen.Get(), nil)
		if err != nil{
			return analyzerList, err
		}
		respParserList := genResponsePasers()
		a, err := analyzer.New(mid, respParserList, module.CalculateScoreSimple)
		if err != nil{
			return analyzerList, err
		}
		analyzerList = append(analyzerList, a)
	}
	return analyzerList, nil
}

func GetPipeline(number uint8, dirPath string)([]module.Pipeline, error){
	pipelineList := []module.Pipeline{}
	if number == 0{
		return pipelineList, nil
	}
	for i:=uint8(0); i<number; i++{
		mid, err := module.GenMID(module.TYPE_PIPELINE, snGen.Get(), nil)
		if err != nil{
			return pipelineList, err
		}
		p, err := pipeline.New(mid, genItemProcessors(dirPath), module.CalculateScoreSimple)
		if err != nil{
			return pipelineList, nil
		}
		p.SetFailFast(true)
		pipelineList = append(pipelineList, p)
	}
	return pipelineList, nil
}