#!/bin/bash
# Usage: ./race-test.sh [BASE_URL] [RESERVE_ENDPOINT]
# Example: ./race-test.sh
#          ./race-test.sh http://localhost:3000
#          ./race-test.sh http://localhost:3000 /api/v1/inventory/reserve/pessimistic

BASE_URL=${1:-"http://localhost:3000"}
RESERVE_PATH=${2:-"/api/v1/inventory/reserve"}
TOTAL_REQUESTS=100
PRODUCT_SKU="SKU-TEST-001"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

RESERVE_URL="${BASE_URL}${RESERVE_PATH}"
echo -e "${YELLOW}🚀 Race Condition Test${NC}"
echo "======================"
echo "   Endpoint: POST $RESERVE_URL"
echo ""

INITIAL_STOCK=$(curl -s "$BASE_URL/api/v1/inventory/stock/$PRODUCT_SKU" | jq -r '.stock')
echo -e "📦 Initial stock: ${GREEN}$INITIAL_STOCK${NC}"

echo -e "\n📨 Sending $TOTAL_REQUESTS concurrent requests (in parallel)..."

RESULTS_FILE=$(mktemp)
trap "rm -f $RESULTS_FILE" EXIT

for i in $(seq 1 $TOTAL_REQUESTS); do
  REQUEST_ID=$(uuidgen)
  (
    STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
      -X POST "$RESERVE_URL" \
      -H "Content-Type: application/json" \
      -d "{\"sku\":\"$PRODUCT_SKU\",\"quantity\":1,\"requestId\":\"$REQUEST_ID\"}")
    echo "$STATUS" >> "$RESULTS_FILE"
  ) &
done

wait
echo -e "\n✅ All requests completed"

SUCCESS=$(grep -c "200" "$RESULTS_FILE" 2>/dev/null || echo 0)
FAIL=$((TOTAL_REQUESTS - SUCCESS))

sleep 1
FINAL_STOCK=$(curl -s "$BASE_URL/api/v1/inventory/stock/$PRODUCT_SKU" | jq -r '.stock')
EXPECTED_STOCK=$((INITIAL_STOCK - SUCCESS))

echo -e "\n📊 Results:"
echo "   Successful: $SUCCESS"
echo "   Failed: $FAIL"
echo "   Expected stock (initial - success): $EXPECTED_STOCK"
echo "   Actual stock:   $FINAL_STOCK"

if [ "$FINAL_STOCK" -eq "$EXPECTED_STOCK" ]; then
  echo -e "${GREEN}✅ No race condition — stock consistent${NC}"
else
  # More stock left than expected = some successful reserves didn't deduct
  LOST=$((FINAL_STOCK - EXPECTED_STOCK))
  echo -e "${RED}❌ Race condition detected! Lost $LOST updates${NC}"
fi
