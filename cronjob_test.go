package cronjob

import (
	logger "github.com/cihub/seelog"
	"github.com/horacedh/cronjob-executor/bean"
	"github.com/horacedh/cronjob-executor/task"
	"github.com/horacedh/cronjob-executor/utils"
	"testing"
)

// Created in 2025-03-18 20:44.
// @author Horace

type DemoTask struct {
}

// Handle 任务处理方法
func (d DemoTask) Handle(params *task.TaskParams) *task.HandlerResult {
	logger.Infof("task handle, params: %v", utils.ToJsonString(params))
	return task.Success()
}

type DemoTask1 struct {
}

// Handle 任务处理方法
func (d DemoTask1) Handle(params *task.TaskParams) *task.HandlerResult {
	logger.Infof("task1 handle, params: %v", utils.ToJsonString(params))
	return task.Success()
}

// TestExecutorClient 测试执行器客户端
func TestExecutorClient(t *testing.T) {
	client := GetExecutorClient(&bean.ExecutorOptions{
		Address: "http://localhost:9527",
		Tenant:  "horace",
		AppName: "go-example-executor",
		AppDesc: "Go示例执行器",
		Tag:     "common",
		SignKey: "7d890a079948b196756rtf5452d2245t",
	})
	client.AddTask(DemoTask{}, bean.TaskOptions{
		Cron: "* * * * * ? ",
		Name: "Go测试任务",
	})
	client.AddTask(DemoTask1{}, bean.TaskOptions{
		Cron: "* * * * * ? ",
		Name: "Go测试任务1",
	})
	client.Start()
}
