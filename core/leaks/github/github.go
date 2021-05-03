package github

import (
	"context"

	"github.com/megamon/core/leaks/stage"
)

//RunGitSearch : main stage for leak search on github
func RunGitSearch(ctx context.Context) {
	var searchStage SearchStage
	searchStage.Init()
	var rl RateLimiter
	rl.Init()

	stage.RunMiddlewareStage(ctx, &searchStage, &rl, 1, 1)
	searchStage.Close()

	var fetchStage FetchStage
	fetchStage.Init()
	stage.RunStage(ctx, &fetchStage, &rl, 1, 1, 2)
	fetchStage.Close()

	UpdateState(stage.FRAGMENTED, stage.NEW, "github")
}

//UpdateState : change state1 -> state2 for all reports of the type
func UpdateState(prev, next string, reportType string) (err error) {
	//TODO: fill function
	return
}
