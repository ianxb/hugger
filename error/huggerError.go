package error

import (
	"bytes"
	"fmt"
	"strings"
)

type errorType string

const (
	ERROR_TYPE_DOWNLOADER  errorType = "downloader error"
	ERROR_TYPE_ANALYSER    errorType = "analyser error"
	ERROR_TYPE_SCHEDULER   errorType = "scheduler error"
	ERROR_TYPE_ITEMCHANNEL errorType = "item channel error"
)

type HuggerError interface {
	Error() string
	Type() errorType
}

type myHuggerError struct {
	errortype    errorType
	errMsg       string
	fullErrorMsg string
}

func NewHuggerError(errorType errorType, errMsg string) HuggerError {
	return &myHuggerError{errortype: errorType, errMsg: errMsg}
}

func (mh *myHuggerError) Error() string {
	if mh.fullErrorMsg == "" {
		mh.genFullErrorMsg()
	}
	return mh.fullErrorMsg
}

func (mh *myHuggerError) Type() errorType {
	return mh.errortype
}

func (mh *myHuggerError) genFullErrorMsg() {
	var buffer bytes.Buffer
	buffer.WriteString("hugger error: ")
	if mh.errortype != "" {
		buffer.WriteString(string(mh.errortype))
		buffer.WriteString(": ")
	}
	buffer.WriteString(mh.errMsg)
	mh.fullErrorMsg = fmt.Sprintf("%s", buffer.String())
	return
}

type IllegalParameterError struct {
	msg string
}

func NewIllegalParError(err string) IllegalParameterError {
	 return IllegalParameterError{
	 	msg: fmt.Sprintf("illegal paramenter: %s", strings.TrimSpace(err)),
	 }
}

func (Ill IllegalParameterError) Error() string{
	return Ill.msg
}