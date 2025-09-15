package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServerAddr      string
	QueueBufferSize int
	WorkersCount    int
	MaxRetries      int
	BaseDelay       time.Duration
	MaxDelay        time.Duration
}

func Load() *Config {
	return &Config{
		ServerAddr:      getEnv("SERVER_ADDR", ":8080"),
		QueueBufferSize: getEnvAsInt("QUEUE_BUFFER_SIZE", 100),
		WorkersCount:    getEnvAsInt("WORKERS_COUNT", 5),
		MaxRetries:      getEnvAsInt("MAX_RETRIES", 3),
		BaseDelay:       time.Duration(getEnvAsInt("BASE_DELAY_MS", 100)) * time.Millisecond,
		MaxDelay:        time.Duration(getEnvAsInt("MAX_DELAY_MS", 5000)) * time.Millisecond,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func LoadEnv(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening .env file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Пропускаем пустые строки и комментарии
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Парсим KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Удаляем кавычки если есть
		value = strings.Trim(value, `"'`)

		// Устанавливаем переменную окружения
		os.Setenv(key, value)
	}

	return scanner.Err()
}
