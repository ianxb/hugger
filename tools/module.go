package tools

var midTemplate = "%s%d|%s"
type MID string
type Type string
type CalCulatorScore func(counts Counts) uint64

type Counts struct {

}

type SummaryStruct struct {
	ID        MID         `json:"id"`
	Called    uint64      `json:"called"`
	Accepted  uint64      `json:"accepted"`
	Completed uint64      `json:"completed"`
	Handling  uint64      `json:"handling"`
	Extra     interface{} `json:"extra"`
}

const (
	TYPE_DOWNLOADER Type = "downloader"
	TYPE_ANALYZER   Type = "analyser"
	TYPE_PIPELINE   Type = "pipeline"
)

var legalTypeLetterMap = map[Type]string{
	TYPE_DOWNLOADER: "D",
	TYPE_ANALYZER:   "A",
	TYPE_PIPELINE:   "P",
}

type Module interface {
	ID() MID                          //get the module id
	Addr() string                     //get the address of the module
	Score() uint64                    //get the score of the module
	SetScore(score uint64)            //set the score of the module
	ScoreCalculator() CalCulatorScore //get the score calculator
	CalledCount() uint64              //get the called times of the module
	AcceptedCount() uint64            //get the times of the success called of the module
	CompletedCount() uint64           //get the success called times of the module
	HandlingNumber() uint64           //get how many called the module is processing
	Counts() Counts                   //get add numbers
	Summary() SummaryStruct           //get the summary of the module
}

type SNGenerator interface {
	Start() uint64      //get the NO.1 serial number
	Max() uint64        //get the max serial number
	Next() uint64       //get the next serial number
	CycleCount() uint64 //get the cycle count
	Get() uint64        //get a serial number and ready for next calling
}

type Registrar interface {
	Register(module Module) (bool, error) //register module
	Unregister(mid MID) (bool, error)     //unregister module
	Get(moduleType Type) (Module, error)  //get a module
	GetAll() map[MID]Module               //get all module
	Clear()                               // clean up all registered module
}