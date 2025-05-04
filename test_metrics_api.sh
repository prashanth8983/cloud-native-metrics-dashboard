#!/bin/bash

# Base URL
API_URL="http://localhost:8080/api/v1/query/range"

# For macOS, using a different approach to calculate dates
NOW=$(date -u +%s)
END_TIME=$(date -u -r $NOW +"%Y-%m-%dT%H:%M:%SZ")
START_TIME=$(date -u -r $(($NOW - 300)) +"%Y-%m-%dT%H:%M:%SZ")  # 300 seconds = 5 minutes

echo "Testing metrics API with different step formats..."
echo "Start time: $START_TIME"
echo "End time: $END_TIME"
echo

# Let's also test with a metric we know exists
# Test 1: Step with seconds suffix
echo "Test 1: Step with seconds suffix (\"60s\")"
curl -s -X POST "$API_URL" \
-H "Content-Type: application/json" \
-d '{
  "query": "up",
  "start": "'"$START_TIME"'",
  "end": "'"$END_TIME"'",
  "step": "60s"
}' | jq '.'
echo

# Test 2: Step with minutes suffix
echo "Test 2: Step with minutes suffix (\"1m\")"
curl -s -X POST "$API_URL" \
-H "Content-Type: application/json" \
-d '{
  "query": "scrape_duration_seconds",
  "start": "'"$START_TIME"'",
  "end": "'"$END_TIME"'",
  "step": "1m"
}' | jq '.'
echo

# Test 3: Simple query
echo "Test 3: Simple query with 30s step"
curl -s -X POST "$API_URL" \
-H "Content-Type: application/json" \
-d '{
  "query": "prometheus_http_requests_total",
  "start": "'"$START_TIME"'",
  "end": "'"$END_TIME"'",
  "step": "30s"
}' | jq '.'
echo
