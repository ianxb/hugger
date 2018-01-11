package PipeLine

import "hugger/tools"

type PipeLine interface {
	tools.Module
	ItemParse() []ProcessItem  //get the item parse function list
	Send(item Item) []error    //send a item to the channel
	FastFail() bool            //once the process failed in a parse function, will stop the process
	SetFastFail(fastfile bool) //set fast fail turn on or off
}

type ProcessItem func(item Item)(result Item, err error)