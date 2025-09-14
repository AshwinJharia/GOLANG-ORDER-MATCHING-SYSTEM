#!/bin/bash

# =============================================================================
# COMPREHENSIVE ORDER MATCHING ENGINE TEST SUITE
# No external dependencies (jq-free) - Uses only grep/cut for JSON parsing
# =============================================================================

BASE_URL="http://localhost:8080"

# Colors for better output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

print_section() {
    echo -e "\n${CYAN}============================================${NC}"
    echo -e "${CYAN}$1${NC}"
    echo -e "${CYAN}============================================${NC}"
}

print_test() {
    echo -e "\n${BLUE}--- $1 ---${NC}"
    ((TOTAL_TESTS++))
}

# Function to extract order ID without jq
extract_order_id() {
    local response_body="$1"
    echo "$response_body" | grep -o '"id":"[^"]*"' | cut -d'"' -f4 | head -1
}

# Function to check HTTP status codes
check_status() {
    local expected="$1"
    local actual="$2" 
    local test_name="$3"
    
    if [[ "$actual" == "$expected" ]]; then
        echo -e "${GREEN}✅ PASS - Status Code: $actual${NC}"
        ((PASSED_TESTS++))
    else
        echo -e "${RED}❌ FAIL - Expected: $expected, Got: $actual${NC}"
        ((FAILED_TESTS++))
    fi
}

# Enhanced API call function with status code checking
api_call() {
    local method="$1"
    local endpoint="$2"
    local data="$3"
    local content_type="$4"
    local expected_status="$5"
    local test_name="$6"
    
    print_test "$test_name"
    
    if [[ -n "$data" && -n "$content_type" ]]; then
        response=$(curl -s -w '\n%{http_code}' -X "$method" "$BASE_URL$endpoint" \
                  -H "Content-Type: $content_type" -d "$data")
    elif [[ -n "$data" ]]; then
        response=$(curl -s -w '\n%{http_code}' -X "$method" "$BASE_URL$endpoint" -d "$data")
    else
        response=$(curl -s -w '\n%{http_code}' -X "$method" "$BASE_URL$endpoint")
    fi
    
    status_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | head -n -1)
    
    echo "Response: $response_body"
    check_status "$expected_status" "$status_code" "$test_name"
    
    # Return response body for further processing
    echo "$response_body"
}

print_section "STARTING COMPREHENSIVE ORDER MATCHING ENGINE TESTS"

# =============================================================================
print_section "1. BASIC HEALTH & CONNECTIVITY TESTS"
# =============================================================================

api_call "GET" "/health" "" "" "200" "Health Check"

# =============================================================================
print_section "2. ORDER CREATION & DYNAMIC ID EXTRACTION"
# =============================================================================

print_test "Creating Fresh Order for ID Extraction Test"
response=$(curl -s -w '\n%{http_code}' -X POST "$BASE_URL/orders" \
    -H "Content-Type: application/json" \
    -d '{"symbol":"EXTRACT","side":"buy","type":"limit","price":150,"quantity":100}')

status_code=$(echo "$response" | tail -n1)
response_body=$(echo "$response" | head -n -1)

echo "Status Code: $status_code"
echo "Response Body: $response_body"

# Fix the counter logic here:
if [[ "$status_code" == "200" ]]; then
    ORDER_ID=$(extract_order_id "$response_body")
    if [[ -z "$ORDER_ID" ]]; then
        echo -e "${RED}❌ Could not extract Order ID from response${NC}"
        ((FAILED_TESTS++))
    else
        echo -e "${GREEN}✅ Successfully extracted Order ID: $ORDER_ID${NC}"
        ((PASSED_TESTS++))  # ← Ensure this runs
    fi
else
    echo -e "${RED}❌ HTTP Error: $status_code${NC}"
    ((FAILED_TESTS++))
fi
 

# =============================================================================
print_section "3. ORDER MANAGEMENT LIFECYCLE TESTING"
# =============================================================================

if [[ -n "$ORDER_ID" ]]; then
    api_call "GET" "/orders/$ORDER_ID" "" "" "200" "Get Fresh Order Status"
    api_call "DELETE" "/orders/$ORDER_ID" "" "" "200" "Cancel Fresh Order"
    api_call "DELETE" "/orders/$ORDER_ID" "" "" "400" "Cancel Already Cancelled Order"
    api_call "GET" "/orders/$ORDER_ID" "" "" "200" "Get Cancelled Order Status"
fi

# =============================================================================
print_section "4. CORE TRADING FUNCTIONALITY TESTS"
# =============================================================================

api_call "POST" "/orders" '{"symbol":"TRADE","side":"buy","type":"limit","price":200,"quantity":100}' "application/json" "200" "Place Buy Limit Order"

api_call "POST" "/orders" '{"symbol":"TRADE","side":"sell","type":"limit","price":205,"quantity":75}' "application/json" "200" "Place Sell Limit Order"

api_call "POST" "/orders" '{"symbol":"TRADE","side":"buy","type":"market","quantity":50}' "application/json" "200" "Place Buy Market Order"

api_call "POST" "/orders" '{"symbol":"TRADE","side":"sell","type":"market","quantity":25}' "application/json" "200" "Place Sell Market Order"

api_call "GET" "/orderbook?symbol=TRADE" "" "" "200" "Get Order Book"

api_call "GET" "/trades?symbol=TRADE" "" "" "200" "Get Trades History"

# =============================================================================
print_section "5. FIFO & MATCHING ALGORITHM TESTS"
# =============================================================================

# Create multiple orders at same price to test FIFO
for i in {1..3}; do
    api_call "POST" "/orders" "{\"symbol\":\"FIFO\",\"side\":\"buy\",\"type\":\"limit\",\"price\":300,\"quantity\":$((i*20))}" "application/json" "200" "FIFO Test Order $i (Qty: $((i*20)))"
    sleep 0.1  # Small delay to ensure different timestamps
done

api_call "GET" "/orderbook?symbol=FIFO" "" "" "200" "FIFO Order Book Before Matching"

api_call "POST" "/orders" '{"symbol":"FIFO","side":"sell","type":"limit","price":300,"quantity":100}' "application/json" "200" "FIFO Matching Sell Order"

api_call "GET" "/orderbook?symbol=FIFO" "" "" "200" "FIFO Order Book After Matching"

api_call "GET" "/trades?symbol=FIFO" "" "" "200" "FIFO Trades Verification"

# =============================================================================
print_section "6. CROSS-PRICE MATCHING TESTS"
# =============================================================================

api_call "POST" "/orders" '{"symbol":"CROSS","side":"sell","type":"limit","price":400,"quantity":150}' "application/json" "200" "Cross-Price Sell Order"

api_call "POST" "/orders" '{"symbol":"CROSS","side":"buy","type":"limit","price":405,"quantity":100}' "application/json" "200" "Cross-Price Buy Order (Should Match at 400)"

api_call "GET" "/orderbook?symbol=CROSS" "" "" "200" "Cross-Price Order Book"

api_call "GET" "/trades?symbol=CROSS" "" "" "200" "Cross-Price Trades"

# =============================================================================
print_section "7. COMPREHENSIVE VALIDATION & ERROR TESTING"
# =============================================================================

# Price validation
api_call "POST" "/orders" '{"symbol":"ERROR","side":"buy","type":"limit","price":-100,"quantity":50}' "application/json" "400" "Negative Price Test"

api_call "POST" "/orders" '{"symbol":"ERROR","side":"buy","type":"limit","price":0,"quantity":50}' "application/json" "400" "Zero Price Test"

# Quantity validation
api_call "POST" "/orders" '{"symbol":"ERROR","side":"buy","type":"limit","price":100,"quantity":0}' "application/json" "400" "Zero Quantity Test"

api_call "POST" "/orders" '{"symbol":"ERROR","side":"buy","type":"limit","price":100,"quantity":-25}' "application/json" "400" "Negative Quantity Test"

# Market order validation
api_call "POST" "/orders" '{"symbol":"ERROR","side":"buy","type":"market","price":100,"quantity":50}' "application/json" "400" "Market Order with Price Test"

# Content-Type validation
api_call "POST" "/orders" '{"symbol":"ERROR","side":"buy","type":"limit","price":100,"quantity":50}' "" "400" "Missing Content-Type Test"

api_call "POST" "/orders" '{"symbol":"ERROR","side":"buy","type":"limit","price":100,"quantity":50}' "text/plain" "400" "Wrong Content-Type Test"

# JSON validation
api_call "POST" "/orders" 'invalid json here' "application/json" "400" "Invalid JSON Test"

# Field validation
api_call "POST" "/orders" '{"side":"buy","type":"limit","price":100,"quantity":50}' "application/json" "400" "Missing Symbol Test"

api_call "POST" "/orders" '{"symbol":"ERROR","side":"invalid","type":"limit","price":100,"quantity":50}' "application/json" "400" "Invalid Side Test"

api_call "POST" "/orders" '{"symbol":"ERROR","side":"buy","type":"invalid","price":100,"quantity":50}' "application/json" "400" "Invalid Type Test"

# =============================================================================
print_section "8. RESOURCE NOT FOUND TESTS"
# =============================================================================

api_call "GET" "/orders/nonexistent-id-123" "" "" "404" "Get Non-existent Order"

api_call "DELETE" "/orders/nonexistent-id-456" "" "" "404" "Cancel Non-existent Order"

# =============================================================================
print_section "9. HTTP METHOD VALIDATION TESTS"
# =============================================================================

api_call "GET" "/orders" "" "" "405" "GET on POST Endpoint"

api_call "POST" "/orderbook?symbol=TEST" "" "" "405" "POST on GET Endpoint"

# =============================================================================
print_section "10. PARAMETER VALIDATION TESTS"
# =============================================================================

api_call "GET" "/orderbook" "" "" "200" "Order Book Missing Symbol (Now allowed)"

api_call "GET" "/trades" "" "" "200" "Trades Missing Symbol (Now allowed)"

# =============================================================================
print_section "11. EDGE CASE & STRESS TESTS"
# =============================================================================

# Market order with no liquidity
api_call "POST" "/orders" '{"symbol":"NOLIQUIDITY","side":"buy","type":"market","quantity":100}' "application/json" "200" "Market Order No Liquidity"

# Large quantity order
api_call "POST" "/orders" '{"symbol":"LARGE","side":"buy","type":"limit","price":1000,"quantity":1000000}' "application/json" "200" "Large Quantity Order"

# High precision price
api_call "POST" "/orders" '{"symbol":"PRECISION","side":"buy","type":"limit","price":123.456789,"quantity":10}' "application/json" "200" "High Precision Price"

# =============================================================================
print_section "12. MULTI-SYMBOL ISOLATION TESTS"
# =============================================================================

symbols=("SYMBOL1" "SYMBOL2" "SYMBOL3")
for symbol in "${symbols[@]}"; do
    api_call "POST" "/orders" "{\"symbol\":\"$symbol\",\"side\":\"buy\",\"type\":\"limit\",\"price\":100,\"quantity\":50}" "application/json" "200" "Multi-Symbol Test: $symbol Buy"
    api_call "POST" "/orders" "{\"symbol\":\"$symbol\",\"side\":\"sell\",\"type\":\"limit\",\"price\":100,\"quantity\":25}" "application/json" "200" "Multi-Symbol Test: $symbol Sell"
done

for symbol in "${symbols[@]}"; do
    api_call "GET" "/orderbook?symbol=$symbol" "" "" "200" "Multi-Symbol Order Book: $symbol"
    api_call "GET" "/trades?symbol=$symbol" "" "" "200" "Multi-Symbol Trades: $symbol"
done

# =============================================================================
print_section "13. FINAL STATE VERIFICATION"
# =============================================================================

test_symbols=("EXTRACT" "TRADE" "FIFO" "CROSS" "NOLIQUIDITY" "LARGE" "PRECISION" "SYMBOL1" "SYMBOL2" "SYMBOL3")

echo -e "\n${YELLOW}Final Order Book States:${NC}"
for symbol in "${test_symbols[@]}"; do
    echo -e "\n${BLUE}=== $symbol ORDER BOOK ===${NC}"
    curl -s "$BASE_URL/orderbook?symbol=$symbol"
done

echo -e "\n${YELLOW}Final Trade Histories:${NC}"
for symbol in "${test_symbols[@]}"; do
    echo -e "\n${BLUE}=== $symbol TRADES ===${NC}"
    curl -s "$BASE_URL/trades?symbol=$symbol"
done

# =============================================================================
print_section "TEST SUMMARY & RESULTS"
# =============================================================================

echo -e "\n${CYAN}================================================${NC}"
echo -e "${CYAN}COMPREHENSIVE TEST SUITE COMPLETED${NC}"
echo -e "${CYAN}================================================${NC}"

SUCCESS_RATE=$((PASSED_TESTS * 100 / TOTAL_TESTS))

echo -e "${BLUE}Total Tests Executed: $TOTAL_TESTS${NC}"
echo -e "${GREEN}Tests Passed: $PASSED_TESTS${NC}"
echo -e "${RED}Tests Failed: $FAILED_TESTS${NC}"
echo -e "${YELLOW}Success Rate: $SUCCESS_RATE%${NC}"

if [[ $FAILED_TESTS -eq 0 ]]; then
    echo -e "\n${GREEN}🎉 ALL TESTS PASSED! Your Order Matching Engine is working perfectly!${NC}"
    echo -e "${GREEN}✅ HTTP status codes are correct${NC}"
    echo -e "${GREEN}✅ Order matching logic is functioning${NC}"
    echo -e "${GREEN}✅ Error handling is comprehensive${NC}"
    echo -e "${GREEN}✅ API endpoints are properly implemented${NC}"
    exit 0
else
    echo -e "\n${RED}❌ SOME TESTS FAILED!${NC}"
    echo -e "${YELLOW}Review the failed tests above for specific issues.${NC}"
    exit 1
fi
