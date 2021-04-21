package stage

import (
	"compress/gzip"
	"context"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"time"
)

func doRequest(req *http.Request) (resp *http.Response, err error) {
	client := http.Client{
		Timeout: time.Duration(5 * time.Second),
	}

	resp, err = client.Do(req)
	return resp, err
}

func getBodyReader(resp *http.Response) (bodyReader io.ReadCloser, err error) {
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		bodyReader, err = gzip.NewReader(resp.Body)

	default:
		bodyReader = resp.Body
	}

	if err != nil {
		resp.Body.Close()
	}

	return bodyReader, err
}

//DoRequests : common part of leak search
func DoRequests(ctx context.Context, stage *Interface, reqQueue chan Request, rl RateLimiter, responses chan Response) {
	reqCount := make(map[int]int)
	var err error
	for r := range reqQueue {

	DOREQUEST:
		for {
			httpResp, rErr := doRequest(r.Req)
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
				logInfo("Skipping " + r.Req.URL.String() + " after " + strconv.Itoa(reqCount[r.ID]) + "attempts")
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
func ProcessResponses(ctx context.Context, stage *Interface, respQueue chan Response) {
	for resp := range respQueue {
		bodyReader, err := getBodyReader(resp.Resp)
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

		err = (*stage).ProcessResponse(body)
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

//RunStage : Main processing function
func RunStage(stage *Interface, limiter RateLimiter, nRequestWorkers, nProcessWorkers int) (err error) {
	reqQueue, err := (*stage).BuildRequests()
	respQueue := make(chan Response, MAXCHANCAP)
	ctx := context.Background()

	if err != nil {
		return
	}

	for i := 0; i < nRequestWorkers; i++ {
		go DoRequests(ctx, stage, reqQueue, limiter, respQueue)
	}

	for i := 0; i < nProcessWorkers; i++ {
		go ProcessResponses(ctx, stage, respQueue)
	}

	return
}
