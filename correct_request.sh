#!/bin/bash
# load_test.sh - Скрипт для массового добавления задач

SERVER="http://localhost:8080"
COUNT=20

echo "Starting load test with $COUNT tasks..."

for i in $(seq 1 $COUNT); do
    # Случайные параметры для разнообразия
    TASK_ID="task_$i_$(date +%s)"
    PAYLOAD="{\\\"index\\\":$i,\\\"timestamp\\\":\\\"$(date)\\\",\\\"data\\\":\\\"test data $i\\\"}"
    RETRIES=$((RANDOM % 5 + 1))
    
    echo "Sending task $i with $RETRIES retries..."
    
    # Правильно формируем JSON с экранированием
    JSON_DATA="{\"id\":\"$TASK_ID\",\"payload\":\"$PAYLOAD\",\"max_retries\":$RETRIES}"
    
    curl -s -X POST "$SERVER/enqueue" \
      -H "Content-Type: application/json" \
      -d "$JSON_DATA" \
      -o /dev/null \
      -w "HTTP Status: %{http_code}, Time: %{time_total}s\n"
    
    # Небольшая задержка между запросами
    sleep 0.1
done

echo "Load test completed!"
echo "Checking health status..."
curl -s "$SERVER/healthz" | python3 -m json.tool