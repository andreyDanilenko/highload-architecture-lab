#!/bin/bash

BASE_URL=${1:-"http://localhost:3000"}
TOTAL_REQUESTS=100
PRODUCT_SKU="SKU-TEST-001"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}🚀 Race Condition Test${NC}"
echo "======================"

INITIAL_STOCK=$(curl -s "$BASE_URL/api/v1/inventory/stock/$PRODUCT_SKU" | jq -r '.stock')
echo -e "📦 Initial stock: ${GREEN}$INITIAL_STOCK${NC}"

echo -e "\n📨 Sending $TOTAL_REQUESTS concurrent requests..."

SUCCESS=0
FAIL=0

for i in $(seq 1 $TOTAL_REQUESTS); do
  # Use uuidgen for REAL UUID v4
  REQUEST_ID=$(uuidgen)
  
  STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST "$BASE_URL/api/v1/inventory/reserve" \
    -H "Content-Type: application/json" \
    -d "{\"sku\":\"$PRODUCT_SKU\",\"quantity\":1,\"requestId\":\"$REQUEST_ID\"}")
  
  if [ "$STATUS" = "200" ]; then
    SUCCESS=$((SUCCESS + 1))
  else
    FAIL=$((FAIL + 1))
  fi
  
  if [ $((i % 10)) -eq 0 ]; then
    echo -n "."
  fi
done

wait
echo -e "\n✅ All requests completed"

sleep 1
FINAL_STOCK=$(curl -s "$BASE_URL/api/v1/inventory/stock/$PRODUCT_SKU" | jq -r '.stock')
EXPECTED_STOCK=$((INITIAL_STOCK - TOTAL_REQUESTS))

echo -e "\n📊 Results:"
echo "   Successful: $SUCCESS"
echo "   Failed: $FAIL"
echo "   Expected stock: $EXPECTED_STOCK"
echo "   Actual stock:   $FINAL_STOCK"

if [ "$FINAL_STOCK" -eq "$EXPECTED_STOCK" ]; then
  echo -e "${GREEN}✅ No race condition detected${NC}"
else
  LOST=$((EXPECTED_STOCK - FINAL_STOCK))
  echo -e "${RED}❌ Race condition detected! Lost $LOST updates${NC}"
fi
