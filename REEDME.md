# Task Queue Service

**Асинхронная очередь задач с фоновыми воркерами на Go.**

## 🚀 Возможности

- Асинхронная обработка задач через буферизированную очередь
- Пулы воркеров с настраиваемым количеством параллельных обработчиков
- Экспоненциальный бэкофф с джиттером для повторных попыток
- Graceful shutdown с завершением текущих задач
- Healthcheck endpoint для мониторинга состояния
- Конфигурация через переменные окружения

---

## 📦 Установка

### Требования

- Go 1.24+
- Linux/macOS/Windows

### Сборка

```bash
git clone <repository-url>
cd task-queue
go build -o bin/server cmd/server/main.go
```

---

## ⚙️ Конфигурация

### Переменные окружения

| Переменная           | По умолчанию | Описание                          |
|----------------------|--------------|-----------------------------------|
| `SERVER_ADDR`        | `:8080`      | Адрес HTTP сервера                |
| `QUEUE_BUFFER_SIZE`  | `64`        | Размер буфера очереди              |
| `WORKERS_COUNT`      | `4`         | Количество воркеров               |
| `MAX_RETRIES`        | `3`          | Максимальное количество повторов  |
| `BASE_DELAY_MS`      | `100`        | Базовая задержка повтора (мс)     |
| `MAX_DELAY_MS`       | `5000`       | Максимальная задержка повтора (мс)|
| `LOG_LEVEL`          | `info`       | Уровень логирования              |

### Пример `.env` файла

```env
SERVER_ADDR=:8080
QUEUE_BUFFER_SIZE=64
WORKERS_COUNT=4
MAX_RETRIES=5
BASE_DELAY_MS=100
MAX_DELAY_MS=5000
LOG_LEVEL=info
```

---

## 🚀 Запуск

### Простой запуск

```bash
#запуск производится из корневой директории
go run cmd/main.go --filename="config.env"
```


## 📡 API Endpoints

### `POST /enqueue`

Добавление задачи в очередь.

**Request:**

```json
{
  "id": "string",
  "payload": "string",
  "max_retries": 3
}
```

**Response:**

- `202 Accepted` — задача принята
- `400 Bad Request` — невалидный запрос
- `503 Service Unavailable` — очередь переполнена или остановлена

**Пример:**

```bash
curl -X POST http://localhost:8080/enqueue \
  -H "Content-Type: application/json" \
  -d '{"id":"task1","payload":"{\"action\":\"process\"}","max_retries":3}'
```

### `GET /healthz`

Проверка состояния сервиса.

**Response:**

```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "queue": {
    "queueLength": 5,
    "activeWorkers": 10,
    "bufferCapacity": 100
  },
  "version": "1.0.0"
}
```

**Пример:**

```bash
curl http://localhost:8080/healthz
```

---

## 🔧 Архитектура

### Компоненты системы

```
HTTP Server → Buffered Channel → Worker Pool → Task Processing
```

- **HTTP Server** — принимает задачи через REST API
- **Buffered Channel** — буферизированная очередь задач
- **Worker Pool** — пул горутин для обработки задач
- **Task Processor** — обработчик бизнес-логики задач

### Обработка задач

```go
// Пример обработчика задачи
func TaskHandler(ctx context.Context, task *queue.Task) error {
    // Симуляция работы (100-500ms)
    processingTime := time.Duration(100+rand.Intn(400)) * time.Millisecond

    select {
    case <-time.After(processingTime):
        // 20% chance of failure для тестирования повторов
        if rand.Float32() < 0.2 {
            return fmt.Errorf("simulated processing error")
        }
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

---

## ⚡ Экспоненциальный бэкофф

При ошибках обработки используется стратегия повторных попыток с экспоненциальным увеличением задержки:

```
Попытка 1: 100ms ± jitter
Попытка 2: 200ms ± jitter
Попытка 3: 400ms ± jitter
...
```

Jitter добавляет случайность для предотвращения синхронизации множества клиентов.

---

## 🛡️ Graceful Shutdown

Сервер корректно обрабатывает сигналы завершения:

- `SIGINT/SIGTERM` — прекращает прием новых задач
- Завершает текущие задачи с таймаутом 30 секунд
- Останавливает воркеры после завершения обработки

```bash
# Корректная остановка
Ctrl+C
# или
kill -TERM <pid>
```

---
### Примеры тестовых запросов

```bash
# Валидный запрос
curl -X POST http://localhost:8080/enqueue \
  -d '{"id":"test1","payload":"{\"action\":\"process\"}","max_retries":3}'

# Невалидный payload
curl -X POST http://localhost:8080/enqueue \
  -d '{"id":"test2","payload":"simple string","max_retries":2}'

# Пропущено обязательное поле
curl -X POST http://localhost:8080/enqueue \
  -d '{"payload":"test","max_retries":1}'
```

---

## 📊 Мониторинг

### Healthcheck metrics

```json
{
  "queueLength": 5,          // Текущая длина очереди
  "activeWorkers": 10,       // Активные воркеры
  "bufferCapacity": 100      // Емкость буфера
}
```

### Логирование

Сервер логирует ключевые события:

- Добавление задач в очередь
- Начало обработки задачи
- Успешное завершение
- Ошибки обработки
- Повторные попытки
- Graceful shutdown

---

## 🐛 Диагностика проблем

| Проблема                     | Решение                                                                 |
|------------------------------|-------------------------------------------------------------------------|
| Очередь переполнена          | Увеличьте `QUEUE_BUFFER_SIZE` или уменьшите нагрузку                   |
| Задачи не обрабатываются     | Проверьте `activeWorkers > 0` через `/healthz`                        |
| Высокая задержка обработки   | Увеличьте `WORKERS_COUNT` или настройте `BASE_DELAY_MS/MAX_DELAY_MS`   |

---