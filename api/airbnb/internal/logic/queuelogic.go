/**
 * @Author $
 * @Description //TODO $
 * @Date $ $
 * @Param $
 * @return $
 **/
package logic

import "github.com/Cosmosmay/ctw_zy/api/airbnb/internal/types"

type TaskQueue struct {
	queue chan types.Task
}

func NewTaskQueue(size int) *TaskQueue {
	return &TaskQueue{
		queue: make(chan types.Task, size),
	}
}

func (q *TaskQueue) Enqueue(task types.Task) {
	q.queue <- task
}

func (q *TaskQueue) Dequeue() types.Task {
	return <-q.queue
}

func (q *TaskQueue) Close() {
	close(q.queue)
}
