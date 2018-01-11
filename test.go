package main

//import "fmt"

type Test interface {
	getResult() string
}

type myTest struct {
	result string
}

func (tt *myTest)getResult() string{
	return tt.result
}

func NewTest(result string) Test{
	return &myTest{result:result}
}