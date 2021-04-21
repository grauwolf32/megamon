package stage

import (
	"context"
	"crypto/sha1"
	"io/ioutil"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/megamon/core/config"
	"github.com/megamon/core/leaks/fragment"
	"github.com/megamon/core/leaks/helpers"
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
func RunMiddlewareStage(stage *MiddlewareInterface, limiter RateLimiter, nRequestWorkers, nProcessWorkers int) (err error) {
	reqQueue, err := (*stage).BuildRequests()
	respQueue := make(chan Response, MAXCHANCAP)
	ctx := context.Background()

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
func RunStage(stage *Interface, limiter RateLimiter, nRequestWorkers, nProcessWorkers int) (err error) {
	middleware := (*stage).(MiddlewareInterface)
	RunMiddlewareStage(&middleware, limiter, nRequestWorkers, nProcessWorkers)
	keywords := config.Settings.LeakGlobals.Keywords

	if len(keywords) == 0 {
		return
	}

	for _, text := range (*stage).GetTextsToProcess() {
		var kwContextFragments [][]fragment.Fragment
		var kwFragments [][]fragment.Fragment

		for _, keyword := range keywords {
			kwFragment := fragment.GetKeywordFragments(text, keyword)
			kwContextFragment := fragment.GetKeywordContext(text, CONTEXTLEN, kwFragment)
			kwContextFragments = append(kwContextFragments, kwContextFragment)
			kwFragments = append(kwFragments)
		}

		mergedContexts := fragment.MergeFragments(kwContextFragments, MAXCONTEXTLEN)

		mergedKeywords := kwFragments[0]
		for i := 1; i < len(kwFragments); i++ {
			mergedKeywords = fragment.Merge(mergedKeywords, kwFragments[i])
		}

		kwInFrags := fragment.GetKeywordsInFragments(mergedKeywords, mergedContexts)
		for id := range kwInFrags {
			var textFragment helpers.TextFragment
			keywords := kwInFrags[id]

			frag := mergedContexts[id]
			fragText, err := frag.Apply(text)
			if err != nil {
				logErr(err)
				continue
			}

			textFragment.Text = fragText
			textFragment.ShaHash = sha1.Sum([]byte(fragText))

			for _, kwID := range keywords {
				kw := mergedKeywords[kwID]
				err = kw.ConvertToRunes(text)
				if err != nil {
					logErr(err)
					continue
				}
				textFragment.Keywords = append(textFragment.Keywords, []int{kw.Offset - frag.Offset, kw.Length})
			}
			
		}

	}

	return
}
