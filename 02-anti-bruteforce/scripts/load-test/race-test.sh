#!/bin/bash
# Usage: ./naive-bruteforce-test.sh [BASE_URL] [PATH] [TOTAL_REQUESTS]
# Example:
#   ./naive-bruteforce-test.sh
#   ./naive-bruteforce-test.sh http://localhost:3000 /login 20
#   ./naive-bruteforce-test.sh http://localhost:3000 /resource/naive 20

BASE_URL=${1:-"http://localhost:3000"}
PATH=${2:-"/login"}
TOTAL_REQUESTS=${3:-20}

URL="${BASE_URL}${PATH}"

echo "🚀 Naive rate-limit test"
echo "========================"
echo "  Endpoint: POST $URL"
echo "  Requests: $TOTAL_REQUESTS"
echo ""

RESULTS_FILE=$(mktemp)
trap "rm -f $RESULTS_FILE" EXIT

for i in $(seq 1 $TOTAL_REQUESTS); do
  (
    STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
      -X POST "$URL")
      # -H "X-Forwarded-For: 1.2.3.4")
    echo "$STATUS" >> "$RESULTS_FILE"
  ) &
done

wait
echo "✅ All requests completed"
echo ""

OK_200=$(grep -F "200" "$RESULTS_FILE" 2>/dev/null | wc -l | tr -d ' ')
RL_429=$(grep -F "429" "$RESULTS_FILE" 2>/dev/null | wc -l | tr -d ' ')
OTHER=$((TOTAL_REQUESTS - OK_200 - RL_429))

echo "📊 Results:"
echo "  200 OK:           $OK_200"
echo "  429 Too Many:     $RL_429"
echo "  Other statuses:   $OTHER"
