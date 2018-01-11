package downloader

import (
	"net/http"
	"hugger/tools"
)

type Request struct {
	httpReq *http.Request
	depth int32
}

type Response struct {
	httpRes *http.Response
	depth int32
}

func NewRequest(httpReq *http.Request, depth int32) *Request{
	return &Request{httpReq:httpReq, depth:depth}
}

func (httpReq *Request)HttpReq() *http.Request{
	return httpReq.httpReq
}

func (req *Request) Depth() int32{
	return req.depth
}

func NewResponse(httpRes *http.Response, depth int32) *Response{
	return &Response{httpRes:httpRes, depth:depth}
}

func (httpRes *Response) HttpRes() *http.Response{
	return httpRes.httpRes
}

func (httpRes *Response) Depth() int32{
	return httpRes.depth
}

type Downloader interface {
	tools.Module
	Download(req *http.Request) (*Response, error)
}