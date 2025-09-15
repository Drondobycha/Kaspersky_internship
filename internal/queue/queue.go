package queue

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

type QueueStats struct {
	QueueLength    int
	ActiveWorkers  int
	BufferCapacity int
}

var (
	ErrQueueFull    = errors.New("queue is full")
	ErrQueueStopped = errors.New("queue is stopped")
)

type Queue struct {
	tasks         chan *Task
	results       chan *TaskResult
	activeWorkers int
	mu            sync.RWMutex
	shutdown      chan struct{}
	wg            sync.WaitGroup
	handler       TaskHandler
}

func NewQueue(bufferSize int, handler TaskHandler) *Queue {
	return &Queue{
		tasks:    make(chan *Task, bufferSize),
		results:  make(chan *TaskResult, bufferSize),
		shutdown: make(chan struct{}),
		handler:  handler,
	}
}

func (q *Queue) EnQueue(task *Task) error {
	select {
	case q.tasks <- task:
		task.Status = StatusQueued
		task.CreatedAt = time.Now()
		return nil
	case <-q.shutdown:
		return ErrQueueStopped
	default:
		return ErrQueueFull
	}
}

func (q *Queue) StartWorkers(ctx context.Context, numWorkers int) {
	q.mu.Lock()
	q.activeWorkers = numWorkers
	q.mu.Unlock()
	for i := 0; i < numWorkers; i++ {
		q.wg.Add(1)
		go q.worker(ctx, i)
	}
}

func (q *Queue) worker(ctx context.Context, id int) {
	defer q.wg.Done()
	log.Printf("Worker: %d start work", id)
	for {
		select {
		case task := <-q.tasks:
			q.processTask(ctx, task)
		case <-ctx.Done():
			return
		case <-q.shutdown:
			return
		}
	}
}

func (q *Queue) processTask(ctx context.Context, task *Task) {
	now := time.Now()
	task.StartedAt = &now
	task.Status = StatusRunning

	log.Printf("Starting processing of task %s (attempt 1/%d)", task.ID, task.MaxRetries+1)

	var err error
	for attempt := 1; attempt <= task.MaxRetries+1; attempt++ {
		if attempt > 1 {
			log.Printf("Retrying task %s (attempt %d/%d)", task.ID, attempt, task.MaxRetries+1)
		}

		err = q.handler(ctx, task)
		if err == nil {
			task.Status = StatusDone
			now := time.Now()
			task.CompletedAt = &now
			log.Printf("Task %s completed successfully after %d attempts", task.ID, attempt)
			q.results <- &TaskResult{Task: task}
			return
		}

		task.RetryCount = attempt - 1
		task.Error = err.Error()

		if attempt > task.MaxRetries {
			log.Printf("Task %s failed after %d attempts: %v", task.ID, attempt, err)
			break
		}

		// Exponential backoff with jitter
		delay := ExponentialBackoffWithJitter(attempt, 100*time.Millisecond, 5*time.Second)
		log.Printf("Task %s failed (attempt %d), retrying in %v: %v", task.ID, attempt, delay, err)

		select {
		case <-time.After(delay):
			// Retry after delay
		case <-ctx.Done():
			task.Status = StatusFailed
			log.Printf("Task %s cancelled due to context cancellation", task.ID)
			q.results <- &TaskResult{Task: task, Error: ctx.Err()}
			return
		case <-q.shutdown:
			task.Status = StatusFailed
			log.Printf("Task %s cancelled due to shutdown", task.ID)
			q.results <- &TaskResult{Task: task, Error: ErrQueueStopped}
			return
		}
	}

	task.Status = StatusFailed
	now = time.Now()
	task.CompletedAt = &now
	log.Printf("Task %s permanently failed: %v", task.ID, err)
	q.results <- &TaskResult{Task: task, Error: err}
}

func (q *Queue) Stop() {
	close(q.shutdown)
	q.wg.Wait()
	close(q.tasks)
	close(q.results)
}

func (q *Queue) GetStats() QueueStats {
	q.mu.RLock()
	defer q.mu.RUnlock()

	return QueueStats{
		QueueLength:    len(q.tasks),
		ActiveWorkers:  q.activeWorkers,
		BufferCapacity: cap(q.tasks),
	}
}

func (q *Queue) IsStopped() bool {
	select {
	case <-q.shutdown:
		return true
	default:
		return false
	}
}
