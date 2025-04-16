package services

import (
	"fmt"
	pq "github.com/emirpasic/gods/queues/priorityqueue"
	"github.com/horacedh/cronjob-executor/task"
	"testing"
)

func TestPriorityQueue(t *testing.T) {
	queue := pq.NewWith(taskComparator) // empty
	queue.Enqueue(task.TaskParams{ExecutionTime: 1})
	queue.Enqueue(task.TaskParams{ExecutionTime: 2})
	queue.Enqueue(task.TaskParams{ExecutionTime: 3})
	queue.Enqueue(task.TaskParams{ExecutionTime: 4})
	queue.Enqueue(task.TaskParams{ExecutionTime: 5})
	queue.Enqueue(task.TaskParams{ExecutionTime: 6})
	queue.Enqueue(task.TaskParams{ExecutionTime: 7})

	fmt.Println(queue)
	fmt.Println(queue.Size())
	value, ok := queue.Dequeue()
	fmt.Println(value, ok)
}
