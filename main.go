package main

import (
	"context"
	"os"

	"github.com/megamon/core/leaks/models"
	"github.com/megamon/core/utils"
	"github.com/megamon/web/backend"
)

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

	/*
		go func() {
			err = github.RunGitSearch(ctx)
			if err != nil {
				utils.ErrorLogger.Fatal(err.Error())
			}
		}()
	*/

	githubTaskQueue := make(chan int)
	gistTaskQueue := make(chan int)

	params := make(map[string](chan int))
	params["github"] = githubTaskQueue
	params["gist"] = gistTaskQueue

	var b backend.Backend
	b.Start(params)
	return
}
