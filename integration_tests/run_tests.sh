#!/bin/bash

set -e

echo "🧪 Running SRS Integration Tests"
echo "================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BINARY_NAME="srs"
TEST_BINARY_NAME="srs-test"
LOG_DIR="./integration_tests/logs"
TIMEOUT="60s"

# Create log directory
mkdir -p "$LOG_DIR"

echo -e "${BLUE}📁 Setting up test environment...${NC}"

# Find the binary to test
if [ -f "./$TEST_BINARY_NAME" ]; then
    BINARY_PATH="./$TEST_BINARY_NAME"
    echo -e "${GREEN}✓ Using test binary: $BINARY_PATH${NC}"
elif [ -f "./$BINARY_NAME" ]; then
    BINARY_PATH="./$BINARY_NAME"
    echo -e "${YELLOW}⚠ Using main binary: $BINARY_PATH${NC}"
else
    echo -e "${RED}❌ No SRS binary found. Please build with: go build -o srs .${NC}"
    exit 1
fi

# Verify binary works
echo -e "${BLUE}🔍 Verifying binary functionality...${NC}"
if ! "$BINARY_PATH" version >/dev/null 2>&1; then
    echo -e "${RED}❌ Binary verification failed${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Binary verification passed${NC}"

# Initialize Go modules for tests
echo -e "${BLUE}📦 Initializing Go modules for tests...${NC}"
cd integration_tests
if [ ! -f "go.mod" ]; then
    go mod init integration_tests
    # Add necessary dependencies based on what we're importing
    go mod edit -require=github.com/stretchr/testify@latest
fi

# Clean up any previous test artifacts
mkdir -p "$LOG_DIR"
rm -rf "$LOG_DIR"/*

echo -e "${BLUE}🚀 Running integration tests...${NC}"

# Test categories
TEST_CATEGORIES=(
    "cli_tests"
    "mcp_tests"
    "workflow"
)

TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to run tests in a category
run_test_category() {
    local category=$1
    echo -e "${BLUE}📝 Running $category tests...${NC}"
    
    if [ "$category" = "workflow" ]; then
        # Workflow tests are in the root
        TEST_PATTERN="./workflow_test.go"
    else
        TEST_PATTERN="./$category/"
    fi
    
    if go test -v "$TEST_PATTERN" 2>&1 | tee "$LOG_DIR/${category}_results.log"; then
        echo -e "${GREEN}✅ $category tests: PASSED${NC}"
        local passed=$(grep -c "^--- PASS:" "$LOG_DIR/${category}_results.log" || echo "0")
        PASSED_TESTS=$((PASSED_TESTS + passed))
    else
        echo -e "${RED}❌ $category tests: FAILED${NC}"
        local failed=$(grep -c "^--- FAIL:" "$LOG_DIR/${category}_results.log" || echo "1") 
        FAILED_TESTS=$((FAILED_TESTS + failed))
    fi
    
    # Count total tests in this category
    local total=$(grep -c "=== RUN" "$LOG_DIR/${category}_results.log" || echo "1")
    TOTAL_TESTS=$((TOTAL_TESTS + total))
}

# Run all test categories
for category in "${TEST_CATEGORIES[@]}"; do
    if [ -d "$category" ] || [ "$category" = "workflow" ]; then
        run_test_category "$category"
        echo
    else
        echo -e "${YELLOW}⚠ Skipping $category (not found)${NC}"
    fi
done

cd ..

# Summary
echo -e "${BLUE}📊 Test Summary${NC}"
echo "=============="
echo -e "Total Tests: ${TOTAL_TESTS}"
echo -e "Passed: ${GREEN}${PASSED_TESTS}${NC}"
echo -e "Failed: ${RED}${FAILED_TESTS}${NC}"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}🎉 All integration tests passed!${NC}"
    exit 0
else
    echo -e "${RED}💥 Some integration tests failed!${NC}"
    echo -e "${YELLOW}📋 Check logs in: $LOG_DIR${NC}"
    exit 1
fi