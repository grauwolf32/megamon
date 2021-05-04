package github

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/megamon/core/utils"
	"golang.org/x/time/rate"
)

func logErr(err error) {
	fmt.Println("[ERROR] " + err.Error())
	utils.ErrorLogger.Println(err.Error())
	return
}

func logInfo(info string) {
	utils.InfoLogger.Println(info)
	return
}

//Init : RateLimiter init function
func (rl *RateLimiter) Init() {
	rl.Limiter = rate.NewLimiter(rate.Every(rl.Duration)*rate.Limit(rl.RequestRate), 1)
}

//Wait until desired rate
func (rl *RateLimiter) Wait(ctx context.Context, resp *http.Response) interface{} {
	rlRemaining := resp.Header.Get("x-ratelimit-remaining")
	rlReset := resp.Header.Get("x-ratelimit-reset")

	var remaining, resetTime int
	var err error

	if rlRemaining != "" {
		remaining, err = strconv.Atoi(rlRemaining)

		if err != nil {
			logErr(err)
			return rl.Limiter.Wait(ctx)
		}

		if remaining > 0 {
			return struct{}{}
		}

		if rlReset != "" {
			resetTime, err = strconv.Atoi(rlReset)
			if err != nil {
				logErr(err)
				return rl.Limiter.Wait(ctx)
			}
			toWait := int64(resetTime) - time.Now().Unix()

			if toWait < 120 {
				<-time.After(time.Duration(toWait) * time.Second)
				return struct{}{}
			}
		}
	}

	return rl.Limiter.Wait(ctx)
}
