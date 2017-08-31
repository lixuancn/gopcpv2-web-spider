package module

import "net/http"

type Module interface {
	//当前组件的mid
	ID() MID
	Addr() string
	//组件的评分
	Score() uint64
	SetScore(uint64)
	//获取评分的计算器
	ScoreCalculator()CalculateScore
	//获取组件被调用的次数
	CalledCount()uint64
	//获取当前组件接收的调用次数，在参数错误或超负荷时会拒绝接收
	AcceptedCount()uint64
	//已成功完成的调用次数
	CompletedCount()uint64
	//正在调用组件的数量
	HandlingNumber()uint64
	//一次性获取所有计数
	Counts()Counts
	//用户获取组件的摘要
	Summary() SummaryStruct
}

// Counts 代表用于汇集组件内部计数的类型。
type Counts struct {
	// CalledCount 代表调用计数。
	CalledCount uint64
	// AcceptedCount 代表接受计数。
	AcceptedCount uint64
	// CompletedCount 代表成功完成计数。
	CompletedCount uint64
	// HandlingNumber 代表实时处理数。
	HandlingNumber uint64
}

type SummaryStruct struct{
	ID MID `json:"id"`
	Called uint64 `json:"id"`
	Accepted uint64 `json:"id"`
	Completed uint64 `json:"id"`
	Handling uint64 `json:"id"`
	//如果extra字段不为空，则解析该字段
	Extra interface{} `json:"extra,omitempty"`
}

type Downloader interface {
	Module
	Download(req *Request)(*Response, error)
}

//分析器，并发安全
type Analyzer interface {
	Module
	//当前分析器所用的响应解析函数的列表
	RespParsers() []ParseResponse
	//解析并返回条目,会对若干个解析后的结果进行合并
	Analyze(resp *Response)([]Data, []error)
}

type ParseResponse func(httpResp *http.Response, respDepth uint32)([]Data, []error)


//条目处理管道，并发安全
type Pipeline interface {
	Module
	//返回当前条目处理管道使用的条目处理函数的列表
	ItemProcessors() []ProcessItem
	//发送条目，条目被发送后会依次经过若干的条目处理函数的处理
	Send()
	//当前条目处理管道是否是快速失败的，快速失败是只要在某条目被处理的某一个步骤上出错，则该条目后续处理都会忽略
	FailFast()bool
	SetFailFast(failFast bool)
}
//接收需要处理的条目，返回处理后的结果
type ProcessItem func(item Item)(Item, error)