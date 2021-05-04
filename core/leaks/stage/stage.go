package stage

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/megamon/core/utils"
)

//DoRequests : common part of leak search
func DoRequests(ctx context.Context, stage MiddlewareInterface, reqQueue chan Request, rl RateLimiter, responses chan Response) {
	reqCount := make(map[int]int)
	for r := range reqQueue {

	DOREQUEST:
		for {
			logInfo(fmt.Sprintf("request to %s; count: %d; id: %d", r.Req.URL.String(), reqCount[r.ID], r.ID))

			httpResp, rErr := utils.DoRequest(r.Req)
			reqCount[r.ID]++

			if rErr != nil {
				//If timeout, check for request count
				if err, ok := rErr.(net.Error); ok && err.Timeout() {
					reqCount[r.ID]++
					if reqCount[r.ID] > MAXRETRIES {
						break DOREQUEST
					}

					logErr(err)
					_ = rl.Wait(ctx, &http.Response{})
					continue

				} else {
					logErr(rErr)
					break DOREQUEST
				}
			}

			_ = rl.Wait(ctx, httpResp)
			resp := Response{r.ID, httpResp}
			check := stage.CheckResponse(resp, reqCount[r.ID])

			switch check {
			case OK:
				responses <- resp
				break DOREQUEST
			case WAIT:
				<-time.After(TIMEWAIT * time.Second)
			case SKIP:
				logInfo("skipping " + r.Req.URL.String() + " after " + strconv.Itoa(reqCount[r.ID]) + " attempts")
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
		}
	}
}

//ProcessResponses : common part of leak search
func ProcessResponses(ctx context.Context, stage MiddlewareInterface, respQueue chan Response) {
	for resp := range respQueue {
		logInfo(fmt.Sprintf("processing response from request: %d", resp.RequesID))

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

		err = stage.ProcessResponse(body, resp.RequesID)
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
func RunMiddlewareStage(ctx context.Context, stage MiddlewareInterface, limiter RateLimiter, nRequestWorkers, nProcessWorkers int) (err error) {
	logInfo("building request queue")
	reqQueue := make(chan Request, MAXCHANCAP)
	respQueue := make(chan Response, MAXCHANCAP)

	var wgRequests sync.WaitGroup
	wgRequests.Add(1)
	go func() {
		defer close(reqQueue)
		defer wgRequests.Done()
		_ = stage.BuildRequests(reqQueue)
		return
	}()

	logInfo("initializing stage request workers")
	for i := 0; i < nRequestWorkers; i++ {
		wgRequests.Add(1)
		go func() {
			defer wgRequests.Done()
			DoRequests(ctx, stage, reqQueue, limiter, respQueue)
		}()
	}

	var wg sync.WaitGroup
	logInfo("initializing stage processing workers")
	for i := 0; i < nProcessWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ProcessResponses(ctx, stage, respQueue)
		}()
	}

	wgRequests.Wait()
	close(respQueue)
	wg.Wait()

	return
}

//RunStage : Main processing function
func RunStage(ctx context.Context, stage Interface, limiter RateLimiter, nRequestWorkers, nProcessWorkers, nFragmentizeWorkers int) (err error) {
	middleware := stage.(MiddlewareInterface)
	RunMiddlewareStage(ctx, middleware, limiter, nRequestWorkers, nProcessWorkers)

	logInfo("fragmentizing reports")
	Fragmentize(ctx, stage, nFragmentizeWorkers)
	return
}
