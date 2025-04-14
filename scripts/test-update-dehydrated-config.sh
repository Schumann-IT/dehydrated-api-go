#!/bin/bash

# Test script for update-dehydrated-config.sh

DEBUG=${1:-false}
WORKSPACE=${WORKSPACE:-$(pwd)/}

RESULT_CODE=0

chmod +x ${WORKSPACE}scripts/update-dehydrated-config.sh

# Function to create a test environment
setup_test_env() {
  local test_dir=$(mktemp -d)
  mkdir -p "$test_dir"
  
  # Create default config
  cp "${WORKSPACE}examples/config/dehydrated" "$test_dir/"

  echo "$test_dir"
}

# Function to run a test
run_test() {
  local test_name="$1"
  local env_vars="$2"
  local expected_error="$3"
  local skip=${4:-0}

  if [ $skip -gt 0 ]; then
    echo "Skipping $test_name"
    return
  fi

  # Create test environment
  local test_dir=$(setup_test_env)

  echo "Running test: $test_name"

  # Run the script with environment variables
  if [ -n "$expected_error" ]; then
    (
      cd "$test_dir"
      CONFIG_FILE="$test_dir/dehydrated" eval "$env_vars ${WORKSPACE}scripts/update-dehydrated-config.sh" > $test_dir/update.log
    ) || true
  else
    (
      cd "$test_dir"
      CONFIG_FILE="$test_dir/dehydrated" eval "$env_vars ${WORKSPACE}scripts/update-dehydrated-config.sh"
    )
  fi

  if [ -n "$expected_error" ]; then
    if grep -q "$expected_error" "$test_dir/update.log"; then
      echo "✅ Test passed"
    else
      echo "❌ Test failed"
      echo "Expected error: $expected_error"
      echo "Actual output:"
      cat "$test_dir/update.log"
      RESULT_CODE=1
    fi
  else
    # Check if the output matches expected settings
    for v in $(echo $env_vars); do
      SETTING=$(echo $v | sed -E 's/^DEHYDRATED_//')
      SETTING_NAME=$(echo $SETTING | awk -F"=" '{print $1}')
      if ! grep -q "$SETTING_NAME" "$test_dir/dehydrated"; then
        MISSING="$SETTING_NAME $MISSING"
        RESULT_CODE=1
      fi
    done

    if [ $RESULT_CODE -eq 0 ]; then
      echo "✅ Test passed"
    else
      echo "❌ Test failed"
      echo "Expected to find: '$MISSING' settings"
      echo "Actual output:"
      cat "$test_dir/dehydrated"
    fi
  fi

  if [ "$DEBUG" == "false" ]; then
    # Clean up test directory
    rm -rf "$test_dir"
  else
    echo "Preserved test dir: $test_dir"
  fi
}

# Test 1: setting base dir throws an error
run_test "Setting base dir throws an error" \
  'DEHYDRATED_BASEDIR=/foo/bar' \
  "DEHYDRATED_BASEDIR must not be set!"

# Test 2: setting domains.txt throws an error
run_test "Setting domains.txt throws an error" \
  'DEHYDRATED_DOMAINS_TXT=/foo/bar/domains.txt' \
  "DEHYDRATED_DOMAINS_TXT must not be set!"

# Test 3: config changes
run_test "multiple config changes with overrides" \
  'DEHYDRATED_CHALLENGETYPE="dns-01" DEHYDRATED_CA="letsencrypt-test" DEHYDRATED_KEY_ALGO="rsa"'

echo

if [ $RESULT_CODE -gt 0 ]; then
  echo "❌ Tests failed"
  exit 1
else
  echo "✅ All tests completed successfully!"
  exit 0
fi