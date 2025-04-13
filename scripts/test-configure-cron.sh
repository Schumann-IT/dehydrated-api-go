#!/bin/bash

# Test script for configure-cron.sh

DEBUG=${1:-false}

RESULT_CODE=0

chmod +x ${WORKSPACE}scripts/configure-cron.sh

# Function to create a test environment
setup_test_env() {
  local test_dir=$(mktemp -d)
  mkdir -p "$test_dir"

  echo "$test_dir"
}

# Function to run a test
run_test() {
  local test_name="$1"
  local env_vars="$2"
  local expected_output="$3"
  local skip=${4:-0}

  if [ $skip -gt 0 ]; then
    echo "Skipping $test_name"
    return
  fi

  # Create test environment
  local test_dir=$(setup_test_env)
  expected_output="$expected_output /app/scripts/renew-certs.sh >> $test_dir/cron.log 2>&1"

  echo "Running test: $test_name"

  # Run the script with environment variables
  (
    cd "$test_dir"
    eval BASE_DIR="$test_dir" CRON_DIR="$test_dir" $env_vars ${WORKSPACE}scripts/configure-cron.sh
  )

  # Check if the output matches expected
  if [ ! -f "$test_dir/renew-certs" ]; then
    if [ -z "$env_vars" ]; then
      echo "✅ Test passed"
    else
      echo "❌ Test failed"
      echo "Expected to find: $test_dir/renew-certs"
    fi
  else
    if grep -q "$expected_output" "$test_dir/renew-certs"; then
      echo "✅ Test passed"
    else
      echo "❌ Test failed"
      echo "Expected to find: $expected_output"
      echo "Actual output:"
      cat "$test_dir/renew-certs"
      RESULT_CODE=1
    fi
  fi

  if [ "$DEBUG" == "false" ]; then
    # Clean up test directory
    rm -rf "$test_dir"
  else
    echo "Preserved test dir: $test_dir"
  fi
}

# Test 1: cron not created for empty schedule
run_test "cron file not created if schedule is empty"

# Test 2: config with schedule
run_test "with default app user" \
  'CRON_SCHEDULE="0 3 3 3 3"' \
  '0 3 3 3 3 root'

# Test 3: config with non default app user
run_test "with non default app user" \
  'APP_USER="example" CRON_SCHEDULE="0 3 3 3 3"' \
  '0 3 3 3 3 example'

echo

if [ $RESULT_CODE -gt 0 ]; then
  echo "❌ Tests failed"
  exit 1
else
  echo "✅ All tests completed successfully!"
  exit 0
fi