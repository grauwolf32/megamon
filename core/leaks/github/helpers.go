package github

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/time/rate"
)

//RateLimiter : rate limit request to github
type RateLimiter struct {
	RequestRate float64
	Duration    time.Duration
	Limiter     *rate.Limiter
}

//Init : RateLimiter init function
func (rl *RateLimiter) Init() {
	rl.Limiter = rate.NewLimiter(rate.Every(rl.Duration)*rate.Limit(rl.RequestRate), 1)
}

//Wait until desired rate
func (rl *RateLimiter) Wait(ctx context.Context, resp *http.Response, t time.Time) interface{} {
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
