package gotask

var global *TaskHandle

func InitTaskHandle(workpoolsize int, workpoolbuflen int) {
	global = NewTaskHandle(workpoolsize, workpoolbuflen)
	global.StartWorkerPool()
}

func GetAsyncTaskHandle() *TaskHandle {
	return global
}

func SendAsyncTaskData(body IBody) {
	global.SendToTaskQueue(body)
}

func RegAsyncTask(tasks ...ITask) {
	for i := range tasks {
		global.AddTask(tasks[i])
	}
}
