#!/bin/bash

# Test script for update-config.sh

DEBUG=${1:-false}
WORKSPACE=${WORKSPACE:-$(pwd)/}

RESULT_CODE=0

chmod +x ${WORKSPACE}scripts/update-config.sh

# Function to create a test environment
setup_test_env() {
  local test_dir=$(mktemp -d)
  mkdir -p "$test_dir"
  
  # Create default config
  cp "${WORKSPACE}examples/config.yaml" "$test_dir/"

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

  echo "Running test: $test_name"

  # Run the script with environment variables
  (
    cd "$test_dir"
    CONFIG_FILE="$test_dir/config.yaml" BASE_DIR="$test_dir" eval "$env_vars ${WORKSPACE}scripts/update-config.sh"
  )

  # Check if the output matches expected
  if grep -q "$expected_output" "$test_dir/config.yaml"; then
    echo "✅ Test passed"
  else
    echo "❌ Test failed"
    echo "Expected to find: $expected_output"
    echo "Actual output:"
    cat "$test_dir/config.yaml"
    RESULT_CODE=1
  fi

  if [ "$DEBUG" == "false" ]; then
    # Clean up test directory
    rm -rf "$test_dir"
  else
    echo "Preserved test dir: $test_dir"
  fi
}

# Test 1: Basic configuration override
run_test "Basic configuration" \
  "PORT=8080 ENABLE_WATCHER=true ENABLE_OPENSSL_PLUGIN=false" \
  "port: 8080"

# Test 2: Single plugin with minimal configuration
run_test "Single plugin minimal config" \
  'EXTERNAL_PLUGINS={"\"azure"\":{"\"path"\":"\"/app/plugins/azure-hook.sh\""}}' \
  "  azure:"

# Test 3: Single plugin with full configuration
run_test "Single plugin full config" \
  'EXTERNAL_PLUGINS={"\"azure"\":{"\"enabled"\":true,"\"path"\":"\"/app/plugins/azure-hook.sh"\","\"config"\":{"\"key"\":"\"value"\"}}}' \
  "  azure:"

# Test 4: Multiple plugins
run_test "Multiple plugins" \
  'EXTERNAL_PLUGINS={"\"azure"\":{"\"path"\":"\"/app/plugins/azure-hook.sh"\"},"\"dns"\":{"\"path"\":"\"/app/plugins/dns-hook.sh"\"}}' \
  "  azure:"

# Test 5: Plugin with missing path
run_test "Plugin with missing path" \
  'EXTERNAL_PLUGINS={"\"azure"\":{}}' \
  "plugins:"

# Test 6: Plugin with disabled status
run_test "Plugin with disabled status" \
  'EXTERNAL_PLUGINS={"\"azure"\":{"\"enabled"\":false,"\"path"\":"\"/app/plugins/azure-hook.sh"\"}}' \
  "    enabled: false"

# Test 7: Plugin with empty config
run_test "Plugin with empty config" \
  'EXTERNAL_PLUGINS={"\"azure"\":{"\"path"\":"\"/app/plugins/azure-hook.sh"\","\"config"\":{}}}' \
  "  azure:"

echo

if [ $RESULT_CODE -gt 0 ]; then
  echo "❌ Tests failed"
  exit 1
else
  echo "✅ All tests completed successfully!"
  exit 0
fi