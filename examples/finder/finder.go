package main

import (
	"flag"
	"fmt"
	"os"
	sched "gopcpv2-web-spider/scheduler"
	"gopcpv2-web-spider/examples/finder/internal"
	"gopcpv2-web-spider/examples/finder/monitor"
	"strings"
	"time"
	"net/http"
)

var firstURL string
var domains string
var depth uint
var dirPath string

func init(){
	flag.StringVar(&firstURL, "first", "http://zhihu.sogou.com/zhihu?query=golang+logo", "请输入入口URL：")
	flag.StringVar(&domains, "domain", "zhihu.com", "请输入允许抓取的HOST列表：")
	flag.UintVar(&depth, "depth", 3, "请输入抓取深度：")
	flag.StringVar(&dirPath, "dir", "./pic", "请输入存放目录：")
}

func Usage(){
	fmt.Fprintf(os.Stderr, "%s的用法：\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "%t finder [flags] \n")
	fmt.Fprintf(os.Stderr, "Flags: \n")
	flag.PrintDefaults()
}

func main(){
	flag.Usage = Usage
	flag.Parse()
	//创建调度器
	scheduler := sched.NewScheduler()
	domainParts := strings.Split(domains, ",")
	acceptedDomains := []string{}
	for _, domain := range domainParts{
		domain = strings.TrimSpace(domain)
		if domain != ""{
			acceptedDomains = append(acceptedDomains, domain)
		}
	}
	dataArgs := sched.DataArgs{
		ReqBufferCap: 50,
		ReqMaxBufferNumber: 1000,
		RespBufferCap: 50,
		RespMaxBufferNumber: 10,
		ItemBufferCap: 50,
		ItemMaxBufferNumber: 100,
		ErrorBufferCap: 50,
		ErrorMaxBufferNumber: 1,
	}
	requestArgs := sched.RequestArgs{
		AcceptedDomains: acceptedDomains,
		MaxDepth: uint32(depth),
	}
	downloader, err := internal.GetDownloaders(1)
	if err != nil{
		fmt.Println("创建下载器组件失败", err.Error())
	}
	analyzers, err := internal.GetAnalyzer(1)
	if err != nil{
		fmt.Println("创建解析器组件失败", err.Error())
	}
	pipelines, err := internal.GetPipeline(1, dirPath)
	if err != nil{
		fmt.Println("创建条目处理器组件失败", err.Error())
	}
	moduleArgs := sched.ModuleArgs{
		Downloaders:downloader,
		Analyzers:analyzers,
		Pipelines:pipelines,
	}
	err = scheduler.Init(requestArgs, dataArgs, moduleArgs)
	if err != nil{
		fmt.Println("初始化调度器失败", err.Error())
	}
	//监控
	checkInterval := time.Second
	summarizeInterval := 100 * time.Microsecond
	maxIdleCount := uint(5)
	checkCountChan := monitor.Monitor(scheduler, checkInterval, summarizeInterval, maxIdleCount, true, internal.Record)
	//启动调度器
	firstHTTPReq, err := http.NewRequest("GET", firstURL, nil)
	if err != nil{
		fmt.Println(err)
		return
	}
	err = scheduler.Start(firstHTTPReq)
	if err != nil{
		fmt.Println("启动调度器发生错误：", err.Error())
		return
	}
	//等待监控结束
	<-checkCountChan
}