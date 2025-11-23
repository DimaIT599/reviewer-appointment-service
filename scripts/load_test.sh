#!/bin/bash

# Load testing script for Reviewer Appointment Service
# Requires: Apache Bench (ab) or hey

BASE_URL="${BASE_URL:-http://localhost:8081}"

echo "=== Load Testing Reviewer Appointment Service ==="
echo "Base URL: $BASE_URL"
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if hey is installed
if command -v hey &> /dev/null; then
    TOOL="hey"
elif command -v ab &> /dev/null; then
    TOOL="ab"
else
    echo -e "${RED}Error: Neither 'hey' nor 'ab' (Apache Bench) is installed${NC}"
    echo "Install hey: go install github.com/rakyll/hey@latest"
    echo "Or install Apache Bench: brew install httpd (macOS) or apt-get install apache2-utils (Linux)"
    exit 1
fi

echo -e "${YELLOW}Setting up test data...${NC}"

# Create a test team
curl -s -X POST "$BASE_URL/team/add" \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "loadtest",
    "members": [
      {"user_id": "lt1", "username": "LoadTest1", "is_active": true},
      {"user_id": "lt2", "username": "LoadTest2", "is_active": true},
      {"user_id": "lt3", "username": "LoadTest3", "is_active": true},
      {"user_id": "lt4", "username": "LoadTest4", "is_active": true},
      {"user_id": "lt5", "username": "LoadTest5", "is_active": true}
    ]
  }' > /dev/null

# Create some PRs
for i in {1..10}; do
  curl -s -X POST "$BASE_URL/pullRequest/create" \
    -H "Content-Type: application/json" \
    -d "{
      \"pull_request_id\": \"pr-load-$i\",
      \"pull_request_name\": \"Load Test PR $i\",
      \"author_id\": \"lt1\"
    }" > /dev/null
done

echo -e "${GREEN}Test data created${NC}"
echo ""

# Test scenarios
echo -e "${YELLOW}=== Test 1: Health Check (High RPS) ===${NC}"
if [ "$TOOL" = "hey" ]; then
    hey -n 1000 -c 10 -m GET "$BASE_URL/health"
else
    ab -n 1000 -c 10 "$BASE_URL/health"
fi
echo ""

echo -e "${YELLOW}=== Test 2: Get Team (Moderate Load) ===${NC}"
if [ "$TOOL" = "hey" ]; then
    hey -n 500 -c 5 -m GET "$BASE_URL/team/get?team_name=loadtest"
else
    ab -n 500 -c 5 "$BASE_URL/team/get?team_name=loadtest"
fi
echo ""

echo -e "${YELLOW}=== Test 3: Get Statistics (Low RPS) ===${NC}"
if [ "$TOOL" = "hey" ]; then
    hey -n 100 -c 2 -m GET "$BASE_URL/statistics"
else
    ab -n 100 -c 2 "$BASE_URL/statistics"
fi
echo ""

echo -e "${YELLOW}=== Test 4: Create PR (Target: 5 RPS) ===${NC}"
if [ "$TOOL" = "hey" ]; then
    hey -n 50 -c 1 -m POST -H "Content-Type: application/json" \
      -d '{"pull_request_id":"pr-load-test","pull_request_name":"Load Test","author_id":"lt2"}' \
      "$BASE_URL/pullRequest/create"
else
    echo "Apache Bench doesn't support POST with body easily, skipping..."
fi
echo ""

echo -e "${GREEN}=== Load Testing Complete ===${NC}"
echo ""
echo "Summary:"
echo "- Health check: High RPS test (1000 requests, 10 concurrent)"
echo "- Get team: Moderate load (500 requests, 5 concurrent)"
echo "- Statistics: Low RPS (100 requests, 2 concurrent)"
echo "- Create PR: Target 5 RPS (50 requests, 1 concurrent)"
echo ""
echo "Expected SLI:"
echo "- Response time: < 300ms (p95)"
echo "- Success rate: > 99.9%"