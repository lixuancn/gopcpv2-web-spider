package module

import "net/http"

type Data interface{
	//判断数据是否有效
	Vaild() bool
}

type Item map[string]interface{}

func (item Item)Vaild()bool{
	return item != nil
}

type Request struct{
	httpReq *http.Request
	depth uint32
}

func NewRequest(httpRequest *http.Request, depth uint32)*Request{
	return &Request{httpReq: httpRequest, depth: depth}
}

func (req *Request)HTTPReq()*http.Request{
	return req.httpReq
}

func (req *Request)Depth()uint32{
	return req.depth
}

func (req *Request)Vaild()bool{
	return req.httpReq != nil && req.httpReq.URL != nil
}

type Response struct{
	httpResp *http.Response
	depth uint32
}

func NewResponse(httpResp *http.Response, depth uint32)*Response{
	return &Response{httpResp:httpResp, depth:depth}
}

func (resp *Response)HTTPResp()*http.Response{
	return resp.httpResp
}

func (resp *Response)Depth()uint32{
	return resp.depth
}

func (resp *Response)Vaild()bool{
	return resp.httpResp != nil && resp.httpResp.Body != nil
}