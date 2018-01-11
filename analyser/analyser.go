package analyser

import (
	"hugger/tools"
	"hugger/downloader"
	"net/http"
)

type ParseResponse func(httpResp *http.Response, resDepth uint64) ([]Data, []error)

type Analyser interface {
	tools.Module
	ResParse() []ParseResponse //get the list of parse function, a response need to be processed by multi parse function
	Analyser(res *downloader.Response) ([]Data, []error)
}
