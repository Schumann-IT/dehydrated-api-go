#!/bin/bash

# This script updates /app/config/dehydrated from environment variables prefixed with DEHYDRATED_

reserved() {
    local value="$1"
    local forbidden=("BASEDIR" "CERTDIR" "DOMAINS_TXT")
    printf '%s\n' "${forbidden[@]}" | grep -q "^${value}$"
}

# the base data dir, variable for testing
CONFIG_FILE=${CONFIG_FILE:-/app/config/dehydrated}

if [ ! -f $CONFIG_FILE ]; then
  touch $CONFIG_FILE
fi

for v in $(env | grep "DEHYDRATED_"); do
  SETTING=$(echo $v | sed -E 's/^DEHYDRATED_//')
  SETTING_NAME=$(echo $SETTING | awk -F'=' '{print $1}')
  if reserved "$SETTING_NAME"; then
    echo "ERROR: DEHYDRATED_$SETTING_NAME must not be set!"
    exit 1
  fi

  # remove the setting if it exists
  sed -i "/^${SETTING_NAME}=/d" $CONFIG_FILE
  echo $SETTING >> $CONFIG_FILE
done
