package main

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/megamon/core/leaks/gist"
	"github.com/megamon/core/leaks/github"
	"github.com/megamon/core/leaks/models"
	"github.com/megamon/core/utils"
	"github.com/megamon/web/backend"
)

func runWorker(taskQueue chan int, task func(ctx context.Context) error) {
	ctx := context.Background()
	var taskID int
	var err error

	for {
		select {
		case taskID = <-taskQueue:
			switch taskID {
			case 1:
				fmt.Printf("Got %d\n", taskID)
				taskQueue <- 2
				err = task(ctx)
				if err != nil {
					utils.ErrorLogger.Println(err.Error())
				}
			default:
				fmt.Printf("Got %d\n", taskID)
			}
		default:
			runtime.Gosched()
		}
	}
}

func main() {
	_ = context.Background()
	utils.InitConfig("./config/config.yaml")

	if _, err := os.Stat(utils.Settings.LeakGlobals.LogDir); os.IsNotExist(err) {
		os.Mkdir(utils.Settings.LeakGlobals.LogDir, os.FileMode(775))
	}

	if _, err := os.Stat(utils.Settings.LeakGlobals.ContentDir); os.IsNotExist(err) {
		os.Mkdir(utils.Settings.LeakGlobals.ContentDir, os.FileMode(775))
	}

	logFilePath := utils.Settings.LeakGlobals.LogDir + utils.Settings.LeakGlobals.LogFile
	utils.InitLoggers(logFilePath)
	utils.InfoLogger.Println("programm started")

	var manager models.Manager
	err := manager.Init()
	if err != nil {
		utils.ErrorLogger.Fatal(err.Error())
		return
	}

	defer manager.Close()
	models.Init(manager.Database)

	params := make(map[string](*utils.WorkerParams))
	params["github"] = &utils.WorkerParams{Task: github.RunGitSearch, Status: utils.TaskNotRunning}
	params["gist"] = &utils.WorkerParams{Task: gist.RunGistStage, Status: utils.TaskNotRunning}

	var b backend.Backend
	b.Start(params)
	return
}
