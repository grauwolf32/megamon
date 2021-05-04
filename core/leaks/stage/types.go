package stage

import (
	"context"
	"fmt"
	"net/http"

	"github.com/megamon/core/leaks/models"
	"github.com/megamon/core/utils"
)

const (
	//OK : response correct
	OK = iota

	//SKIP : skip this request
	SKIP

	//WAIT : try request again
	WAIT
)

const (
	//MAXRETRIES : max number of requests to probe
	MAXRETRIES = 3

	//TIMEWAIT : time to wait after failed request
	TIMEWAIT = 5

	//MAXCHANCAP : max channel capacity
	MAXCHANCAP = 4096

	//CONTEXTLEN : desired length of keyword context
	CONTEXTLEN = 480

	//MAXCONTEXTLEN : max length of keyword context
	MAXCONTEXTLEN = 640
)

const (
	//PROCESSED : first stage of leak processing
	PROCESSED = "processed"

	//FETCHED : second stage; load all content
	FETCHED = "fetched"

	//FRAGMENTED : third stage; generated all fragments
	FRAGMENTED = "fragmented"

	//NEW : fourth stage; report can be viewed
	NEW = "new"

	//CLOSED : report closed
	CLOSED = "closed"

	//VALIDATED : report validated (leak found)
	VALIDATED = "validated"
)

//Request : basic request type
type Request struct {
	ID  int
	Req *http.Request
}

//Response : basic response type
type Response struct {
	RequesID int
	Resp     *http.Response
}

//ReportText : text with report ID to fragmentize
type ReportText struct {
	ReportID int
	Text     string
}

//MiddlewareInterface common pipeline
type MiddlewareInterface interface {
	Init() (err error)
	Close()
	GetDBManager() models.Manager
	BuildRequests(res chan Request) (err error)
	CheckResponse(resp Response, reqCount int) (res int)
	ProcessResponse(resp []byte, RequesID int) (err error)
}

//Interface : common pipeline
type Interface interface {
	MiddlewareInterface
	GetTextsToProcess(chan ReportText) (err error)
	ProcessTextFragment(fragment models.TextFragment) error
}

//RateLimiter : limits requests rate
type RateLimiter interface {
	Wait(ctx context.Context, resp *http.Response) interface{}
}

func logErr(err error) {
	fmt.Println("[ERROR] " + err.Error())
	utils.ErrorLogger.Println(err.Error())
	return
}

func logInfo(info string) {
	utils.InfoLogger.Println(info)
	return
}
