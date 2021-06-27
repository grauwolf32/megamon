package utils

import (
	"context"
	"sync"
)

const (
	//TaskNotRunning : task currently not running, and hasn't been ran yet
	TaskNotRunning = "not_running"

	//TaskCanceled : task was canceled
	TaskCanceled = "canceled"

	//TaskRunning : task is running
	TaskRunning = "running"

	//TaskDone : task done
	TaskDone = "done"
)

//WorkerParams : auxilary structure to hold all needed values
type WorkerParams struct {
	Task      func(ctx context.Context) error
	Status    string
	Cancel    context.CancelFunc
	WaitGroup sync.WaitGroup
}

//RunTask : start task
func RunTask(wp *WorkerParams) {
	if (*wp).Status == TaskRunning {
		return
	}

	var ctx context.Context
	ctx, (*wp).Cancel = context.WithCancel(context.Background())
	(*wp).Status = TaskRunning
	(*wp).WaitGroup.Add(1)

	go func() {
		defer (*wp).WaitGroup.Done()
		err := (*wp).Task(ctx)
		if err != nil {
			ErrorLogger.Println(err.Error())
		}

		(*wp).Status = TaskDone
	}()
	return
}

//EndTask : send cancel message to the running task
func EndTask(wp *WorkerParams) {
	if (*wp).Status != TaskRunning {
		return
	}

	(*wp).Cancel()
	(*wp).WaitGroup.Wait()
	(*wp).Status = TaskCanceled
}
