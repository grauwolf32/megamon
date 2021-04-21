package stage

import (
	"compress/gzip"
	"context"
	"io"
	"io/ioutil"
	"net"
	"net/http"
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
func DoRequests(ctx context.Context, stage Interface, reqQueue chan request, rl *RateLimiter) (responses chan Response) {
	responses = make(chan Response, 4096)
	reqCount := make(map[int]int)
	var err error

	go func() {
		defer responses.Close()
		for r := range reqQueue {

		DOREQUEST:
			for {
				httpResp, rErr := doRequest(r)
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
				check := stage.CheckResponse(resp, reqCount[r.ID])

				if check == OK {
					responses <- resp
				} else if check == SKIP {
					logInfo("Skipping " + r.URL.String() + " after " + reqCount[r.ID] + "attempts")
					break DOREQUEST

				} else if check == WAIT {
					<-time.After(TIMEWAIT * time.Second)
				}

				select {
				case <-ctx.Done():
					return
				default:
				}
			}

			_ = rl.Wait(resp, time.Now())
		}
	}()

	return responses
}

//ProcessResponses : common part of leak search
func ProcessResponses(stage *Interface, respQueue chan Response) {
	for resp := range respQueue {
		bodyReader, err := getBodyReader(resp)
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

		err = stage.ProcessResponse(body)
		if err != nil {
			logErr(err)
		}
	}
}

//RunStage : Main processing function
func RunStage(stage *Interface, limiter RateLimiter) (err error) {
	jobQueue, err := (*stage).BuildRequests()
	if err != nil {
		return
	}

	if err = (*stage).DoRequests(jobQueue, limiter); err != nil {
		return err
	}

	if err = (*stage).ProcessResponses(); err != nil {
		return err
	}

	if err = (*stage).Fragmentize(); err != nil {
		return err
	}
	return
}
