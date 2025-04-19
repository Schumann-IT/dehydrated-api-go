#!/bin/bash

# This script updates the /app/config/config.yaml file from environment variables
# use CONFIG_FILE=/test/dir/config.yaml BASE_DIR=/test/dir ./update-api-config.sh for testing

set -e

# Check if yq is installed
if ! command -v yq &> /dev/null; then
  echo "Error: yq is not installed. Please install it first."
  echo ""
  echo "Installation instructions:"
  echo ""
  echo "For macOS:"
  echo "  brew install yq"
  echo ""
  echo "For Linux:"
  echo "  wget https://github.com/mikefarah/yq/releases/download/v4.40.5/yq_linux_amd64 -O /usr/bin/yq && chmod +x /usr/bin/yq"
  echo ""
  echo "For other platforms, visit: https://github.com/mikefarah/yq#install"
  exit 1
fi

# the base data dir, variable for testing
CONFIG_FILE=${CONFIG_FILE:-/app/config/config.yaml}
# the dehydrated base dir, variable for testing
BASE_DIR=${BASE_DIR:-/data/dehydrated}

# set dehydrated base dir
echo "Setting dehydratedBaseDir to $BASE_DIR"
yq -i ".dehydratedBaseDir = \"$BASE_DIR\"" "$CONFIG_FILE"

# Override specific values if environment variables are set
if [ -n "$PORT" ]; then
  echo "Setting port to $PORT"
  yq -i ".port = $PORT" "$CONFIG_FILE"
fi

if [ -n "$ENABLE_WATCHER" ]; then
  echo "Setting enableWatcher to $ENABLE_WATCHER"
  yq -i ".enableWatcher = $ENABLE_WATCHER" "$CONFIG_FILE"
fi

if [ -n "$ENABLE_OPENSSL_PLUGIN" ]; then
  echo "Setting openssl plugin enabled to $ENABLE_OPENSSL_PLUGIN"
  yq -i ".plugins.openssl.enabled = $ENABLE_OPENSSL_PLUGIN" "$CONFIG_FILE"
fi

# Process external plugins if provided
if [ -n "$EXTERNAL_PLUGINS" ]; then
  echo "Processing external plugins configuration..."
  
  # Create a temporary file for the plugins configuration
  TEMP_PLUGINS_FILE=$(mktemp)

  # Parse external plugin config
  echo $EXTERNAL_PLUGINS | yq -P '.' > $TEMP_PLUGINS_FILE

  # Merge the plugins configuration into the main config
  yq -i ".plugins = load(\"$TEMP_PLUGINS_FILE\")" "$CONFIG_FILE"
  
  # Clean up temporary files
  rm -f "$TEMP_PLUGINS_FILE"
  
  echo "External plugins configuration processed successfully"
fi

echo "Configuration file updated at $CONFIG_FILE"
