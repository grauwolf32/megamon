package stage

import (
	"context"
	"io/ioutil"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/megamon/core/utils"
)

//DoRequests : common part of leak search
func DoRequests(ctx context.Context, stage *MiddlewareInterface, reqQueue chan Request, rl RateLimiter, responses chan Response) {
	reqCount := make(map[int]int)
	var err error
	for r := range reqQueue {

	DOREQUEST:
		for {
			httpResp, rErr := utils.DoRequest(r.Req)
			reqCount[r.ID]++

			if rErr != nil {
				//If timeout, check for request count
				if err, ok := err.(net.Error); ok && err.Timeout() {
					if reqCount[r.ID] > MAXRETRIES {
						break DOREQUEST
					}
					<-time.After(TIMEWAIT * time.Second)

				} else {
					logErr(err)
					break DOREQUEST
				}
			}

			resp := Response{r.ID, httpResp}
			check := (*stage).CheckResponse(resp, reqCount[r.ID])

			switch check {
			case OK:
				responses <- resp
			case WAIT:
				<-time.After(TIMEWAIT * time.Second)
			case SKIP:
				logInfo("Skipping " + r.Req.URL.String() + " after " + strconv.Itoa(reqCount[r.ID]) + " attempts")
				break DOREQUEST
			default:
				if reqCount[r.ID] > MAXRETRIES {
					break DOREQUEST
				}
				<-time.After(TIMEWAIT * time.Second)
			}

			select {
			case <-ctx.Done():
				return
			default:
			}

			_ = rl.Wait(resp.Resp, time.Now())

		}
	}
}

//ProcessResponses : common part of leak search
func ProcessResponses(ctx context.Context, stage *MiddlewareInterface, respQueue chan Response) {
	for resp := range respQueue {
		bodyReader, err := utils.GetBodyReader(resp.Resp)
		if err != nil {
			logErr(err)
			continue
		}

		body, err := ioutil.ReadAll(bodyReader)
		bodyReader.Close()

		if err != nil {
			logErr(err)
			continue
		}

		err = (*stage).ProcessResponse(body, resp.RequesID)
		if err != nil {
			logErr(err)
		}

		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

//RunMiddlewareStage : Middleware processing function
func RunMiddlewareStage(ctx context.Context, stage *MiddlewareInterface, limiter RateLimiter, nRequestWorkers, nProcessWorkers int) (err error) {
	reqQueue, err := (*stage).BuildRequests()
	respQueue := make(chan Response, MAXCHANCAP)

	var wgRequests sync.WaitGroup
	var wg sync.WaitGroup

	if err != nil {
		return
	}

	for i := 0; i < nRequestWorkers; i++ {
		wgRequests.Add(1)
		go func() {
			defer wgRequests.Done()
			DoRequests(ctx, stage, reqQueue, limiter, respQueue)
		}()
	}

	for i := 0; i < nProcessWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ProcessResponses(ctx, stage, respQueue)
		}()
	}

	wgRequests.Wait()
	close(reqQueue)
	wg.Wait()
	return
}

//RunStage : Main processing function
func RunStage(ctx context.Context, stage *Interface, limiter RateLimiter, nRequestWorkers, nProcessWorkers, nFragmentizeWorkers int) (err error) {
	middleware := (*stage).(MiddlewareInterface)
	RunMiddlewareStage(ctx, &middleware, limiter, nRequestWorkers, nProcessWorkers)
	Fragmentize(ctx, stage, nFragmentizeWorkers)
	return
}
