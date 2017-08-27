package errors

import (
	"strings"
	"bytes"
	"fmt"
)

type ErrorType string

const ERROR_TYPE_DOWNLOAD ErrorType = "下载失败"
const ERROR_TYPE_ANALYZER ErrorType = "分析失败"
const ERROR_TYPE_PIPLINE ErrorType = "条目处理管道失败"
const ERROR_TYPE_SCHEDULER ErrorType = "调度器失败"

type CrawlerError interface{
	Type() ErrorType
	Error() string
}

type myCrawlerError struct{
	errType ErrorType
	errMsg string
	fullErrMsg string
}

func NewCrawlerError(errType ErrorType, errMsg string)CrawlerError{
	return &myCrawlerError{
		errType:errType,
		errMsg:strings.TrimSpace(errMsg),
	}
}

func (ce *myCrawlerError)Type()ErrorType{
	return ce.errType
}

func (ce *myCrawlerError)Error()string{
	if ce.fullErrMsg == ""{
		ce.genFullErrMsg()
	}
	return ce.fullErrMsg
}

func (ce *myCrawlerError)genFullErrMsg(){
	var buffer bytes.Buffer
	buffer.WriteString("Crawler Error:")
	if ce.errType != ""{
		buffer.WriteString("【")
		buffer.WriteString(ce.errMsg)
		buffer.WriteString("】: ")
	}
	buffer.WriteString(ce.errMsg)
	ce.fullErrMsg = buffer.String()
}

// IllegalParameterError 代表非法的参数的错误类型。
type IllegalParameterError struct {
	msg string
}

// NewIllegalParameterError 会创建一个IllegalParameterError类型的实例。
func NewIllegalParameterError(errMsg string) IllegalParameterError {
	return IllegalParameterError{
		msg: fmt.Sprintf("参数错误: %s", strings.TrimSpace(errMsg)),
	}
}

func (ipe IllegalParameterError) Error() string {
	return ipe.msg
}