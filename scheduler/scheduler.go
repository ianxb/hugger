package scheduler

import (
	"net/http"
	"hugger/analyser"
	"hugger/downloader"
	"hugger/PipeLine"
	//"hugger/tools"
	"hugger/tools"
)

type Status uint8
type Scheduler interface {
	Init(requestArgs RequestArgs, dataArgs DataArgs, moduleArgs ModuleArgs) (err error)
	Start(firstHttpReq *http.Request) (err error)
	Stop() (err error)
	Status() Status
	ErrorChan() <-chan error
	Idle() bool
	Summary() SchedSummary
}

type RequestArgs struct {
	AcceptedDomains []string `json:"accepted domains"`
	MaxDepth        uint64   `json:"max depth"`
}

type DataArgs struct {
	ReqBufferCap     uint64 `json:"request buffer capacity"`
	ReqBufferNumber  uint64 `json:"request buffer number"`
	ResBufferCap     uint64 `json:"response buffer capacity"`
	ResBufferNumber  uint64 `json:"response buffer number"`
	ItemBufferCap    uint64 `json:"item buffer capacity"`
	ItemBufferNumber uint64 `json:"item buffer number"`
}

type ModuleArgs struct {
	DownLoaders []downloader.Downloader
	Analyser    []analyser.Analyser
	Pipeline    []PipeLine.PipeLine
}

type Args interface {
	Check() error
}

const (
	UNINITIALIZED Status = 0
	INITIALING Status = 1
	INITIALIZED Status = 2
	STARTING Status = 3
	STARTED Status = 4
	STOPPING Status = 5
	STOPPED Status = 6
)

type SchedSummary interface {
	Struct() SummaryStruct //get struct summary info
	String() string//get string summary info
}

type SummaryStruct struct {
	RequestArgs     RequestArgs             `json:"request args"`
	DataArgs        DataArgs                `json:"data args"`
	ModuleArgs      ModuleArgs              `json:"module args"`
	Status          string
	Downloaders     []tools.SummaryStruct   `json:"downloaders"`
	Analyser        []tools.SummaryStruct   `json:"analyser"`
	PipeLine        []tools.SummaryStruct   `json:"pipeline"`
	ReqBufferPool   BufferPoolSummaryStruct `json:"request buffer pool"`
	ItemBufferPool  BufferPoolSummaryStruct `json:"item buffer pool"`
	ErrorBufferPool BufferpoolSummaryStruct `json:"error buffer pool"`
	UrlNum          uint64                  `json:"url number"`
}
