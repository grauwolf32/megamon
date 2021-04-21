package stage

import (
	"fmt"
	"net/http"
	"time"

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

//Interface common pipeline
type Interface interface {
	BuildRequests() (res chan Request, err error)
	CheckResponse(resp Response, reqCount int) (res int)
	ProcessResponse(resp []byte) (err error)
}

//RateLimiter : limits requests rate
type RateLimiter interface {
	Wait(resp *http.Response, t time.Time) interface{}
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
