package scheduler

import (
	"net/http"
	"SafelyMap"
	"gopcpv2-web-spider/module"
	"gopcpv2-web-spider/toolkit/buffer"
	"context"
	"sync"
	"log"
	"fmt"
	"errors"
	"strings"
)

type Scheduler interface {
	Init(requestArgs RequestArgs, dataArgs DataArgs, moduleArgs ModuleArgs)error
	Start(firstHttpReq *http.Request)error
	Stop()error
	Status()Status
	ErrorChan()<-chan error
	Idle()bool
	Summary()SchedSummary
}

type myScheduler struct {
	maxDepth uint32
	acceptedDomainMap SafelyMap.ConcurrentMap
	registrar module.Registrar
	reqBufferPool buffer.Pool
	respBufferPool buffer.Pool
	itemBufferPool buffer.Pool
	errorBufferPool buffer.Pool
	urlMap SafelyMap.ConcurrentMap
	ctx context.Context
	cancelFunc context.CancelFunc
	status Status
	statusLock sync.RWMutex
	summary SchedSummary
}

func NewScheduler()Scheduler{
	return &myScheduler{}
}

func(sched *myScheduler)Init(requestArgs RequestArgs, dataArgs DataArgs, moduleArgs ModuleArgs)(err error){
	log.Printf("检测状态参数")
	oldStatus, err := sched.checkAndSetStatus(SCHED_STATUS_INITIALIZING)
	if err != nil{
		return
	}
	defer func(){
		sched.statusLock.Lock()
		if err != nil{
			sched.status = oldStatus
		}else{
			sched.status = SCHED_STATUS_INITIALIZED
		}
		sched.statusLock.Unlock()
	}()
	log.Println("检测请求参数")
	if err = requestArgs.Check(); err != nil{
		return err
	}
	log.Println("检测数据参数")
	if err = dataArgs.Check(); err != nil{
		return err
	}
	log.Println("检测组件参数")
	if err = moduleArgs.Check(); err != nil{
		return err
	}
	log.Println("初始化内部字段")
	if sched.registrar == nil{
		sched.registrar = module.NewRegistrar()
	}else{
		sched.registrar.Clear()
	}
	sched.maxDepth = requestArgs.MaxDepth
	sched.acceptedDomainMap, _ = SafelyMap.NewConcurrentMap(1, nil)
	for _, domain := range requestArgs.AcceptedDomains{
		sched.acceptedDomainMap.Put(domain, struct {}{})
	}
	sched.urlMap, _ = SafelyMap.NewConcurrentMap(16, nil)
	sched.initBufferPool(dataArgs)
	sched.resetContext()
	sched.summary = newSchedSummary(requestArgs, dataArgs, moduleArgs, sched)
	log.Println("注册组件...")
	if err = sched.registerModules(moduleArgs); err != nil{
		return err
	}
	log.Println("初始化完成")
	return nil
}

func(sched *myScheduler)Start(firstHTTPReq *http.Request)(err error){
	defer func(){
		if p := recover(); p!=nil{
			err = genError(fmt.Sprintf("调度器错误：%s", p))
		}
	}()
	oldStatus, err := sched.checkAndSetStatus(SCHED_STATUS_STARTING)
	defer func(){
		sched.statusLock.Lock()
		if err != nil{
			sched.status = oldStatus
		}else{
			sched.status = SCHED_STATUS_STARTED
		}
		sched.statusLock.Unlock()
	}()
	if err != nil{
		return
	}
	if firstHTTPReq == nil{
		return genParameterError("第一个HTTP请求参数是nil")
	}
	primaryDomain, err := getPrimaryDomain(firstHTTPReq.Host)
	if err != nil{
		return
	}
	sched.acceptedDomainMap.Put(primaryDomain, struct {}{})
	if err = sched.checkBufferPoolForStart(); err != nil{
		return
	}
	sched.download()
	sched.analyze()
	sched.pick()
	log.Println("调度器已经启动")
	firstReq := module.NewRequest(firstHTTPReq, 0)
	sched.sendReq(firstReq)
	return nil
}

func(sched *myScheduler)Stop()error{
	log.Println("关闭调度器")
	oldStatus, err := sched.checkAndSetStatus(SCHED_STATUS_STOPPING)
	defer func(){
		sched.statusLock.Lock()
		if err != nil{
			sched.status = oldStatus
		}else{
			sched.status = SCHED_STATUS_STOPPED
		}
		sched.statusLock.Unlock()
	}()
	if err != nil{
		return err
	}
	sched.cancelFunc()
	sched.reqBufferPool.Close()
	sched.respBufferPool.Close()
	sched.itemBufferPool.Close()
	sched.errorBufferPool.Close()
	log.Println("调度器已关闭")
	return nil
}

func(sched *myScheduler)Status()Status{
	sched.statusLock.RLock()
	defer sched.statusLock.RUnlock()
	status := sched.status
	return status
}

func(sched *myScheduler)ErrorChan()<-chan error {
	errBuffer := sched.errorBufferPool
	errCh := make(chan error, errBuffer.BufferCap())
	go func(errBuffer buffer.Pool, errCh chan error) {
		for {
			if sched.canceled(){
				close(errCh)
				break
			}
			datum ,err := errBuffer.Get()
			if err != nil{
				log.Println("错误缓冲池已关闭")
				close(errCh)
				break
			}
			err, ok := datum.(error)
			if !ok{
				sendError(errors.New(fmt.Sprintf("无效的错误类型：%T", datum)), "", sched.errorBufferPool)
				continue
			}
			if sched.canceled(){
				close(errCh)
				break
			}
			errCh <- err
		}
	}(errBuffer, errCh)
	return errCh
}

func(sched *myScheduler)Idle()bool{
	moduleList := sched.registrar.GetAll()
	for _, modulex := range moduleList{
		if modulex.HandlingNumber() > 0{
			return false
		}
	}
	if sched.reqBufferPool.Total() > 0{
		return false
	}
	if sched.respBufferPool.Total() > 0{
		return false
	}
	if sched.itemBufferPool.Total() > 0{
		return false
	}
	return true
}

func(sched *myScheduler)Summary()SchedSummary{
	return sched.summary
}

func(sched *myScheduler)checkAndSetStatus(status Status)(Status, error){
	sched.statusLock.Lock()
	defer sched.statusLock.Unlock()
	oldStatus := sched.status
	err := checkStatus(oldStatus, status, nil)
	if err == nil{
		sched.status = status
	}
	return oldStatus, err
}

// initBufferPool 用于按照给定的参数初始化缓冲池。
// 如果某个缓冲池可用且未关闭，就先关闭该缓冲池。
func (sched *myScheduler)initBufferPool(dataArgs DataArgs){
	// 初始化请求缓冲池。
	if sched.reqBufferPool != nil && !sched.reqBufferPool.Closed(){
		sched.reqBufferPool.Close()
	}
	sched.reqBufferPool, _ = buffer.NewPool(dataArgs.ReqBufferCap, dataArgs.ReqMaxBufferNumber)
	// 初始化响应缓冲池。
	if sched.respBufferPool != nil && sched.respBufferPool.Closed(){
		sched.respBufferPool.Close()
	}
	sched.respBufferPool, _ = buffer.NewPool(dataArgs.RespBufferCap, dataArgs.RespMaxBufferNumber)
	// 初始化条目缓冲池。
	if sched.itemBufferPool != nil && !sched.itemBufferPool.Closed(){
		sched.itemBufferPool.Close()
	}
	sched.itemBufferPool, _ = buffer.NewPool(dataArgs.ItemBufferCap, dataArgs.ItemMaxBufferNumber)
	// 初始化错误缓冲池。
	if sched.errorBufferPool != nil && !sched.errorBufferPool.Closed(){
		sched.errorBufferPool.Close()
	}
	sched.errorBufferPool, _ = buffer.NewPool(dataArgs.ErrorBufferCap, dataArgs.ErrorMaxBufferNumber)
}

// resetContext 用于重置调度器的上下文。
func (sched *myScheduler) resetContext() {
	sched.ctx, sched.cancelFunc = context.WithCancel(context.Background())
}

// registerModules 会注册所有给定的组件。
func (sched *myScheduler) registerModules(moduleArgs ModuleArgs) error {
	log.Println("注册下载器..")
	for _, d := range moduleArgs.Downloaders {
		if d == nil {
			continue
		}
		ok, err := sched.registrar.Register(d)
		if err != nil {
			return genErrorByError(err)
		}
		if !ok {
			errMsg := fmt.Sprintf("Couldn't register downloader instance with MID %q!", d.ID())
			return genError(errMsg)
		}
	}
	log.Println("注册分析器..")
	for _, a := range moduleArgs.Analyzers {
		if a == nil {
			continue
		}
		ok, err := sched.registrar.Register(a)
		if err != nil {
			return genErrorByError(err)
		}
		if !ok {
			errMsg := fmt.Sprintf("Couldn't register analyzer instance with MID %q!", a.ID())
			return genError(errMsg)
		}
	}
	log.Println("注册通道..")
	for _, p := range moduleArgs.Pipelines {
		if p == nil {
			continue
		}
		ok, err := sched.registrar.Register(p)
		if err != nil {
			return genErrorByError(err)
		}
		if !ok {
			errMsg := fmt.Sprintf("Couldn't register pipeline instance with MID %q!", p.ID())
			return genError(errMsg)
		}
	}
	return nil
}


// checkBufferPoolForStart 会检查缓冲池是否已为调度器的启动准备就绪。
// 如果某个缓冲池不可用，就直接返回错误值报告此情况。
// 如果某个缓冲池已关闭，就按照原先的参数重新初始化它。
func (sched *myScheduler) checkBufferPoolForStart() error {
	// 检查请求缓冲池。
	if sched.reqBufferPool == nil {
		return genError("nil request buffer pool")
	}
	if sched.reqBufferPool.Closed() {
		sched.reqBufferPool, _ = buffer.NewPool(sched.reqBufferPool.BufferCap(), sched.reqBufferPool.MaxBufferNumber())
	}
	// 检查响应缓冲池。
	if sched.respBufferPool == nil {
		return genError("nil response buffer pool")
	}
	if sched.respBufferPool.Closed() {
		sched.respBufferPool, _ = buffer.NewPool(sched.respBufferPool.BufferCap(), sched.respBufferPool.MaxBufferNumber())
	}
	// 检查条目缓冲池。
	if sched.itemBufferPool == nil {
		return genError("nil item buffer pool")
	}
	if sched.itemBufferPool.Closed() {
		sched.itemBufferPool, _ = buffer.NewPool(sched.itemBufferPool.BufferCap(), sched.itemBufferPool.MaxBufferNumber())
	}
	// 检查错误缓冲池。
	if sched.errorBufferPool == nil {
		return genError("nil error buffer pool")
	}
	if sched.errorBufferPool.Closed() {
		sched.errorBufferPool, _ = buffer.NewPool(sched.errorBufferPool.BufferCap(), sched.errorBufferPool.MaxBufferNumber())
	}
	return nil
}

// canceled 用于判断调度器的上下文是否已被取消。
func (sched *myScheduler) canceled() bool {
	select {
	case <-sched.ctx.Done():
		return true
	default:
		return false
	}
}

//会从请求缓冲池取出请求并下载，然后把得到的响应放入响应缓冲池。
func(sched *myScheduler)download(){
	go func(){
		for{
			if sched.canceled(){
				break
			}
			datum, err := sched.reqBufferPool.Get()
			if err != nil{
				break
			}
			req, ok := datum.(*module.Request)
			if !ok{
				sendError(errors.New(fmt.Sprintf("无效的请求类型：%T", datum)), "", sched.errorBufferPool)
			}
			sched.downloadOne(req)
		}
	}()
}

//会根据给定的请求执行下载并把响应放入响应缓冲池。
func(sched *myScheduler)downloadOne(req *module.Request){
	if req == nil || sched.canceled(){
		return
	}
	m, err := sched.registrar.Get(module.TYPE_DOWNLOADER)
	if err != nil || m == nil{
		sendError(errors.New(fmt.Sprintf("获取下载器组件失败：%s", err)), "", sched.errorBufferPool)
		sched.sendReq(req)
		return
	}
	downloader, ok := m.(module.Downloader)
	if !ok{
		sendError(errors.New(fmt.Sprintf("无效的下载器类型：%Y, MID: %s", m, m.ID())), m.ID(), sched.errorBufferPool)
		sched.sendReq(req)
		return
	}
	resp, err := downloader.Download(req)
	if resp != nil{
		sendResp(resp, sched.respBufferPool)
	}
	if err != nil{
		sendError(err, m.ID(), sched.errorBufferPool)
	}
}

// analyze 会从响应缓冲池取出响应并解析，
// 然后把得到的条目或请求放入相应的缓冲池。
func (sched *myScheduler) analyze() {
	go func() {
		for {
			if sched.canceled() {
				break
			}
			datum, err := sched.respBufferPool.Get()
			if err != nil {
				log.Println("从响应缓冲池获取响应失败")
				break
			}
			resp, ok := datum.(*module.Response)
			if !ok {
				sendError(errors.New(fmt.Sprintf("无效的响应类型: %T", datum)), "", sched.errorBufferPool)
			}
			sched.analyzeOne(resp)
		}
	}()
}

// analyzeOne 会根据给定的响应执行解析并把结果放入相应的缓冲池。
func (sched *myScheduler) analyzeOne(resp *module.Response) {
	if resp == nil {
		return
	}
	if sched.canceled() {
		return
	}
	m, err := sched.registrar.Get(module.TYPE_ANALYZER)
	if err != nil || m == nil {
		sendError(errors.New(fmt.Sprintf("获取解析组件失败: %s", err)), "", sched.errorBufferPool)
		sendResp(resp, sched.respBufferPool)
		return
	}
	analyzer, ok := m.(module.Analyzer)
	if !ok {
		sendError(errors.New(fmt.Sprintf("无效的解析器组件类型: %T (MID: %s)", m, m.ID())), m.ID(), sched.errorBufferPool)
		sendResp(resp, sched.respBufferPool)
		return
	}
	dataList, errs := analyzer.Analyze(resp)
	if dataList != nil {
		for _, data := range dataList {
			if data == nil {
				continue
			}
			switch d := data.(type) {
			case *module.Request:
				sched.sendReq(d)
			case module.Item:
				sendItem(d, sched.itemBufferPool)
			default:
				sendError(errors.New(fmt.Sprintf("不支持的数据类型：%T, data:%#v", d, d)), m.ID(), sched.errorBufferPool)
			}
		}
	}
	if errs != nil {
		for _, err := range errs {
			sendError(err, m.ID(), sched.errorBufferPool)
		}
	}
}

// pick 会从条目缓冲池取出条目并处理。
func (sched *myScheduler) pick() {
	go func() {
		for {
			if sched.canceled() {
				break
			}
			datum, err := sched.itemBufferPool.Get()
			if err != nil {
				log.Println("从条目缓存池获取数据失败")
				break
			}
			item, ok := datum.(module.Item)
			if !ok {
				sendError(errors.New(fmt.Sprintf("无效的条目类型: %T", datum)), "", sched.errorBufferPool)
			}
			sched.pickOne(item)
		}
	}()
}

// pickOne 会处理给定的条目。
func (sched *myScheduler) pickOne(item module.Item) {
	if sched.canceled() {
		return
	}
	m, err := sched.registrar.Get(module.TYPE_PIPELINE)
	if err != nil || m == nil {
		sendError(errors.New(fmt.Sprintf("获取条目处理器失败: %s", err)), "", sched.errorBufferPool)
		sendItem(item, sched.itemBufferPool)
		return
	}
	pipeline, ok := m.(module.Pipeline)
	if !ok {
		sendError(errors.New(fmt.Sprintf("无效的条目类型: %T (MID: %s)", m, m.ID())), m.ID(), sched.errorBufferPool)
		sendItem(item, sched.itemBufferPool)
		return
	}
	errs := pipeline.Send(item)
	if errs != nil {
		for _, err := range errs {
			sendError(err, m.ID(), sched.errorBufferPool)
		}
	}
}

func(sched *myScheduler)sendReq(req *module.Request)bool{
	if req == nil{
		return false
	}
	if sched.canceled(){
		return false
	}
	httpReq := req.HTTPReq()
	if httpReq == nil{
		return false
	}
	if httpReq.URL == nil{
		return false
	}
	scheme := strings.ToLower(httpReq.URL.Scheme)
	if scheme != "http" && scheme != "https"{
		return false
	}
	if v := sched.urlMap.Get(httpReq.URL.String()); v != nil{
		return false
	}
	pd, _ := getPrimaryDomain(httpReq.URL.Host)
	if sched.acceptedDomainMap.Get(pd) == nil{
		if pd == "bing.net"{
			panic(httpReq.URL)
		}
		return false
	}
	if req.Depth() > sched.maxDepth{
		return false
	}
	go func(req *module.Request){
		if err := sched.reqBufferPool.Put(req); err != nil{
			log.Println("请求发送给请求缓冲器失败")
		}
	}(req)
	sched.urlMap.Put(httpReq.URL.String(), struct {}{})
	return true
}

func sendResp(resp *module.Response, respBufferPool buffer.Pool) bool {
	if resp == nil || respBufferPool == nil || respBufferPool.Closed() {
		return false
	}
	go func(resp *module.Response) {
		if err := respBufferPool.Put(resp); err != nil {
			log.Println("响应写入响应缓存池失败")
		}
	}(resp)
	return true
}

// sendItem 会向条目缓冲池发送条目。
func sendItem(item module.Item, itemBufferPool buffer.Pool) bool {
	if item == nil || itemBufferPool == nil || itemBufferPool.Closed() {
		return false
	}
	go func(item module.Item) {
		if err := itemBufferPool.Put(item); err != nil {
			log.Println("条目写入条目缓存池失败")
		}
	}(item)
	return true
}