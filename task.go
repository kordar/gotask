package gotask

import (
	logger "github.com/kordar/gologger"
)

type IBody interface {
	TaskId() string
}

type ITask interface {
	Id() string
	Execute(body IBody)
}

type TaskHandle struct {
	Name             string
	Container        map[string]ITask // 存放每个MsgId 所对应的处理方法的map属性
	WorkerPoolSize   int              // 业务工作Worker池的数量
	TaskQueueBuffLen int              // 最大任务长度
	TaskQueue        []chan IBody     // Worker负责取任务的消息队列
	MsgId            int              // 消息id，每次投递递增
}

func NewTaskHandle(workSize int, queueBuffLen int) *TaskHandle {
	return NewTaskHandleWithName("gotask", workSize, queueBuffLen)
}

func NewTaskHandleWithName(name string, workSize int, queueBuffLen int) *TaskHandle {
	return &TaskHandle{
		Container:        make(map[string]ITask),
		WorkerPoolSize:   workSize,
		TaskQueueBuffLen: queueBuffLen,
		Name:             name,
		TaskQueue:        make([]chan IBody, workSize),
	}
}

// SendToTaskQueue 将消息交给TaskQueue,由worker进行处理
func (mh *TaskHandle) SendToTaskQueue(body IBody) {
	// 根据ConnID来分配当前的连接应该由哪个worker负责处理
	// 轮询的平均分配法则

	//得到需要处理此条连接的workerID
	workerID := mh.MsgId % mh.WorkerPoolSize
	//将请求消息发送给任务队列
	mh.TaskQueue[workerID] <- body
	mh.refreshMsgId()
}

// SendToTaskQueueP 将消息交给TaskQueue,由worker进行处理
func (mh *TaskHandle) SendToTaskQueueP(body IBody, pools []int) {
	// 根据ConnID来分配当前的连接应该由哪个worker负责处理
	// 轮询的平均分配法则
	index := mh.MsgId % len(pools)
	//得到需要处理此条连接的workerID
	workerID := pools[index]
	//将请求消息发送给任务队列
	mh.TaskQueue[workerID] <- body
	mh.refreshMsgId()
}

// SendToTaskQueueN 将消息交给TaskQueue,由worker进行处理
func (mh *TaskHandle) SendToTaskQueueN(body IBody, workerID int) {
	if workerID >= mh.WorkerPoolSize {
		return
	}
	mh.TaskQueue[workerID] <- body
	mh.refreshMsgId()
}

func (mh *TaskHandle) refreshMsgId() {
	if mh.MsgId > 1000000 {
		mh.MsgId = 0
	} else {
		mh.MsgId++
	}
}

// AddTask 为消息添加具体的处理逻辑
func (mh *TaskHandle) AddTask(task ITask) {
	// 1 判断当前msg绑定的API处理方法是否已经存在
	taskId := task.Id()
	if _, ok := mh.Container[taskId]; ok {
		panic("repeated func , taskId = " + taskId)
	}
	// 2 添加msg与api的绑定关系
	mh.Container[task.Id()] = task
	logger.Infof("[%s] the task named '%s' was added successfully.", mh.Name, taskId)
}

// DoMsgHandler 马上以非阻塞方式处理消息
func (mh *TaskHandle) DoMsgHandler(body IBody) {
	handler, ok := mh.Container[body.TaskId()]
	if !ok {
		logger.Infof("[%s] no task named '%s' was found..", mh.Name, body.TaskId())
		return
	}

	// 执行对应处理方法
	handler.Execute(body)
}

// StartOneWorker 启动一个Worker工作流程
func (mh *TaskHandle) StartOneWorker(workerID int, taskQueue chan IBody) {
	logger.Infof("[%s] WorkerID = %d is starting.", mh.Name, workerID)
	// 不断的等待队列中的消息
	for {
		select {
		// 有消息则取出队列的Request，并执行绑定的业务方法
		case request := <-taskQueue:
			mh.DoMsgHandler(request)
		}
	}
}

// StartWorkerPool 启动worker工作池
func (mh *TaskHandle) StartWorkerPool() {
	// 遍历需要启动worker的数量，依此启动
	for i := 0; i < mh.WorkerPoolSize; i++ {
		// 一个worker被启动
		// 给当前worker对应的任务队列开辟空间
		mh.TaskQueue[i] = make(chan IBody, mh.TaskQueueBuffLen)
		// 启动当前Worker，阻塞的等待对应的任务队列是否有消息传递进来
		go mh.StartOneWorker(i, mh.TaskQueue[i])
	}
}
