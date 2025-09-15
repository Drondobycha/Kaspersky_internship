package queue

import (
	"context"
	"math"
	"math/rand"
	"time"
)

type TaskStatus string

const (
	StatusQueued  TaskStatus = "queued"
	StatusRunning TaskStatus = "running"
	StatusDone    TaskStatus = "done"
	StatusFailed  TaskStatus = "failed"
)

type Task struct {
	ID          string     `json:"id"`
	Payload     string     `json:"payload"`
	MaxRetries  int        `json:"max_retries"`
	Status      TaskStatus `json:"status"`
	RetryCount  int        `json:"retry_count"`
	Error       string     `json:"error,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type TaskResult struct {
	Task    *Task
	Error   error
	Retried bool
}

type TaskHandler func(ctx context.Context, task *Task) error

// ExponentialBackoffWithJitter calculates delay with exponential backoff and jitter
func ExponentialBackoffWithJitter(attempt int, baseDelay time.Duration, maxDelay time.Duration) time.Duration {
	if attempt <= 0 {
		return baseDelay
	}

	exponent := attempt - 1
	if exponent > 10 {
		exponent = 10
	}

	exponential := float64(baseDelay) * math.Pow(2, float64(exponent))

	jitterFactor := 0.5 + rand.Float64() // 0.5 to 1.5
	delay := time.Duration(exponential * jitterFactor)

	// Cap at max delay
	if delay > maxDelay {
		return maxDelay
	}

	// Ensure minimum delay is at least baseDelay for attempt 1+
	if attempt > 1 && delay < baseDelay {
		return baseDelay
	}

	return delay
}
