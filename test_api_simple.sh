#!/bin/bash

BASE_URL="http://localhost:8080"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

print_test() {
  echo -e "\n${BLUE}--- $1 ---${NC}"
}

check_status() {
  local expected="$1"
  local actual="$2"

  if [[ "$actual" == "$expected" ]]; then
    echo -e "${GREEN}✅ PASS - Status Code: $actual${NC}"
  else
    echo -e "${RED}❌ FAIL - Expected: $expected, Got: $actual${NC}"
  fi
}

# Health Check
print_test "Health Check"
response=$(curl -s -w '\n%{http_code}' $BASE_URL/health)
status=$(echo "$response" | tail -n1)
response_body=$(echo "$response" | head -n -1)
echo "Response: $response_body"
check_status 200 "$status"

# Place Buy Limit Order
print_test "Place Buy Limit Order"
response=$(curl -s -w '\n%{http_code}' -X POST $BASE_URL/orders -H "Content-Type: application/json" \
  -d '{"symbol":"AAPL","side":"buy","type":"limit","price":150.00,"quantity":100}')
status=$(echo "$response" | tail -n1)
response_body=$(echo "$response" | head -n -1)
echo "Response: $response_body"
check_status 200 "$status"

# Place Sell Limit Order (should match)
print_test "Place Sell Limit Order"
response=$(curl -s -w '\n%{http_code}' -X POST $BASE_URL/orders -H "Content-Type: application/json" \
  -d '{"symbol":"AAPL","side":"sell","type":"limit","price":149.00,"quantity":50}')
status=$(echo "$response" | tail -n1)
response_body=$(echo "$response" | head -n -1)
echo "Response: $response_body"
check_status 200 "$status"

# Check Order Book for AAPL
print_test "Check Order Book with Symbol"
response=$(curl -s -w '\n%{http_code}' "$BASE_URL/orderbook?symbol=AAPL")
status=$(echo "$response" | tail -n1)
response_body=$(echo "$response" | head -n -1)
echo "Response: $response_body"
check_status 200 "$status"

# Check Order Book without symbol (should now allow and return all)
print_test "Check Order Book without Symbol"
response=$(curl -s -w '\n%{http_code}' "$BASE_URL/orderbook")
status=$(echo "$response" | tail -n1)
response_body=$(echo "$response" | head -n -1)
echo "Response: $response_body"
check_status 200 "$status"

# Check Trades for AAPL
print_test "Check Trades with Symbol"
response=$(curl -s -w '\n%{http_code}' "$BASE_URL/trades?symbol=AAPL")
status=$(echo "$response" | tail -n1)
response_body=$(echo "$response" | head -n -1)
echo "Response: $response_body"
check_status 200 "$status"

# Check Trades without Symbol (should now allow and return all)
print_test "Check Trades without Symbol"
response=$(curl -s -w '\n%{http_code}' "$BASE_URL/trades")
status=$(echo "$response" | tail -n1)
response_body=$(echo "$response" | head -n -1)
echo "Response: $response_body"
check_status 200 "$status"
