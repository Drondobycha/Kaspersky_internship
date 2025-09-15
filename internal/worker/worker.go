package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"Kaspersky_internship/internal/queue"
)

func TaskHandler(ctx context.Context, task *queue.Task) error {
	//Симулируем обработку полезной нагрузки
	if err := validatePayload(task.Payload); err != nil {
		return fmt.Errorf("payload validation failed: %v", err)
	}

	// Генерируем случайное время работы от 100 до 500 мс
	processingTime := time.Duration(100+rand.Intn(400)) * time.Millisecond

	log.Printf("Processing task %s, estimated time: %v", task.ID, processingTime)

	// Симулируем работу в течение случайного времени
	select {
	case <-time.After(processingTime):
		// Имитируем 20% вероятность ошибки для тестирования повторных попыток
		if rand.Float32() < 0.2 {
			return fmt.Errorf("simulated processing error for task %s (after %v)", task.ID, processingTime)
		}

		log.Printf("Task %s completed successfully in %v", task.ID, processingTime)
		return nil

	case <-ctx.Done():
		return ctx.Err()
	}
}

// validatePayload проверяет валидность JSON payload
func validatePayload(payload string) error {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(payload), &data); err != nil {
		return fmt.Errorf("invalid JSON payload: %v", err)
	}

	// Дополнительные проверки payload при необходимости
	if len(payload) > 1024*1024 { // 1MB limit
		return fmt.Errorf("payload too large")
	}

	return nil
}
