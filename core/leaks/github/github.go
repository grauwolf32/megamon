package github

import (
	"context"

	"github.com/megamon/core/leaks/models"
	"github.com/megamon/core/leaks/stage"
)

//RunGitSearch : main stage for leak search on github
func RunGitSearch(ctx context.Context) (err error) {
	var searchStage SearchStage
	searchStage.Init()
	var rl RateLimiter
	rl.Init()

	logInfo("search stage started")
	err = stage.RunMiddlewareStage(ctx, &searchStage, &rl, 1, 1)
	searchStage.Close()

	if err != nil {
		logErr(err)
		return
	}

	var fetchStage FetchStage
	fetchStage.Init()
	logInfo("fetch stage started")
	err = stage.RunStage(ctx, &fetchStage, &rl, 1, 1, 2)
	fetchStage.Close()

	if err != nil {
		logErr(err)
		return
	}

	err = UpdateState(stage.FRAGMENTED, stage.NEW, "github")
	if err != nil {
		logErr(err)
	}

	return
}

//UpdateState : change state1 -> state2 for all reports of the type
func UpdateState(prev, next string, reportType string) (err error) {
	var manager models.Manager
	err = manager.Init()
	if err != nil {
		return
	}

	query := "UPDATE " + models.ReportTable + " SET status=$2 WHERE status=$1 AND type=$3;"
	_, err = manager.Database.Exec(query, next, prev, reportType)
	return
}
