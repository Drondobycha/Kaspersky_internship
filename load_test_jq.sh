#!/bin/bash
# load_test_jq.sh - Более надежный вариант с jq

SERVER="http://localhost:8080"
COUNT=100

echo "Starting load test with $COUNT tasks..."

for i in $(seq 1 $COUNT); do
    TASK_ID="task_$i_$(date +%s)"
    RETRIES=$((RANDOM % 5 + 1))
    
    # Создаем payload и request с помощью jq
    REQUEST_JSON=$(jq -n \
      --arg id "$TASK_ID" \
      --argjson index "$i" \
      --arg timestamp "$(date)" \
      --arg data "test data $i" \
      --argjson retries "$RETRIES" \
      '{
        id: $id,
        payload: ({"index": $index, "timestamp": $timestamp, "data": $data} | tostring),
        max_retries: $retries
      }')
    
    echo "Sending task $i with $RETRIES retries..."
    
    curl -s -X POST "$SERVER/enqueue" \
      -H "Content-Type: application/json" \
      -d "$REQUEST_JSON" \
      -o /dev/null \
      -w "HTTP Status: %{http_code}, Time: %{time_total}s\n"
    
    sleep 0.1
done

echo "Load test completed!"
curl -s "$SERVER/healthz" | python3 -m json.tools