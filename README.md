# gotask

一个轻量级的 Go 异步任务分发/处理组件：通过 Worker Pool 将任务投递到队列，并按任务 ID 查找对应处理器执行。

## 特性

- Worker Pool + 任务队列（按轮询将任务分配到不同 worker）
- 任务注册（按字符串 TaskId/Id 进行路由）
- 支持三种投递方式：轮询、指定 worker、从指定 pools 中轮询

## 安装

```bash
go get github.com/kordar/gotask
```

## 快速开始

直接使用 `TaskHandle`：

```go
package main

import (
	"fmt"
	"time"

	"github.com/kordar/gotask"
)

type DemoBody struct {
	A int
	B string
}

func (d DemoBody) TaskId() string { return "demo-task" }

type DemoTask struct{}

func (d DemoTask) Id() string { return "demo-task" }

func (d DemoTask) Execute(body gotask.IBody) {
	b := body.(DemoBody)
	fmt.Println("handle:", b.A, b.B)
}

func main() {
	handle := gotask.NewTaskHandle(3, 100)
	handle.StartWorkerPool()
	handle.AddTask(DemoTask{})

	body := DemoBody{A: 100, B: "hello"}
	handle.SendToTaskQueue(body)
	handle.SendToTaskQueue(body)

	time.Sleep(1 * time.Second)
}
```

使用全局异步句柄（适合应用级单例）：

```go
package main

import (
	"fmt"
	"time"

	"github.com/kordar/gotask"
)

type DemoBody struct {
	A int
	B string
}

func (d DemoBody) TaskId() string { return "demo-task" }

type DemoTask struct{}

func (d DemoTask) Id() string { return "demo-task" }

func (d DemoTask) Execute(body gotask.IBody) {
	b := body.(DemoBody)
	fmt.Println("handle:", b.A, b.B)
}

func main() {
	gotask.InitTaskHandle(3, 100)
	gotask.RegAsyncTask(DemoTask{})
	gotask.SendAsyncTaskData(DemoBody{A: 1, B: "async"})

	time.Sleep(1 * time.Second)
}
```

## 核心概念

- `IBody`：任务数据载体，需要实现 `TaskId() string`
- `ITask`：任务处理器，需要实现 `Id() string` 与 `Execute(body IBody)`
- `TaskHandle`：注册任务 + 启动 worker + 投递任务

## 常用 API

- 创建与启动
  - `NewTaskHandle(workSize, queueBuffLen int) *TaskHandle`
  - `(*TaskHandle).StartWorkerPool()`
- 注册任务
  - `(*TaskHandle).AddTask(task ITask)`
  - `RegAsyncTask(tasks ...ITask)`
- 投递任务
  - `(*TaskHandle).SendToTaskQueue(body IBody)`：按轮询分配到 worker
  - `(*TaskHandle).SendToTaskQueueP(body IBody, pools []int)`：在指定 pools 中轮询
  - `(*TaskHandle).SendToTaskQueueN(body IBody, workerID int)`：投递到指定 worker
  - `SendAsyncTaskData(body IBody)`：投递到全局句柄

## 注意事项

- 需要先 `StartWorkerPool()`（或 `InitTaskHandle()`）再投递任务，否则队列未初始化会导致阻塞/异常。
- `SendToTaskQueue*` 内部使用自增计数做轮询分配；若在多个 goroutine 中并发调用投递方法，可能触发数据竞争。建议在外部串行投递，或自行在调用侧加锁/原子控制。

