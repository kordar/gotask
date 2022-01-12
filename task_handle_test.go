package gotask

import (
	"fmt"
	"testing"
	"time"
)

type DemoBody struct {
	A int
	B string
}

func (d DemoBody) TaskId() string {
	return "demo-task"
}

type DemoTask struct {
}

func (d DemoTask) Id() string {
	return "demo-task"
}

func (d DemoTask) Execute(body IBody) {
	demoBody := body.(DemoBody)
	fmt.Println("================", demoBody)
	// log.Println(fmt.Printf("this is demo body %+v", demoBody.B))
}

func TestNewTaskHandle(t *testing.T) {
	handle := NewTaskHandle(3, 200, 100)
	handle.StartWorkerPool()
	handle.AddTask(DemoTask{})

	body := DemoBody{
		A: 100,
		B: "hello",
	}

	handle.SendToTaskQueue(body)
	handle.SendToTaskQueue(body)
	handle.SendToTaskQueue(body)
	handle.SendToTaskQueue(body)
	handle.SendToTaskQueue(body)
	handle.SendToTaskQueue(body)

	time.Sleep(10 * time.Second)
}
