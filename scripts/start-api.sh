#!/bin/sh

CONFIG_FILE=/app/config/config.yaml

# create base dir
mkdir -p /data/dehydrated

# Use config from base dir if exists
if [ ! -f "/data/dehydrated/config.yaml" ]; then
  CONFIG_FILE=/data/dehydrated/config.yaml
fi

# Create empty domains.txt if it does not exist
if [ ! -f "/data/dehydrated/domains.txt" ]; then
  echo "Creating empty domains.txt"
  touch "/data/dehydrated/domains.txt"
fi

# update api config from env variables
/app/scripts/update-api-config.sh
# update dehydrated config from env variables
/app/scripts/update-dehydrated-config.sh
# prepare account
/app/scripts/dehydrated --register --accept-terms --config /app/config/dehydrated
# configure cron if needed
/app/scripts/configure-cron.sh
# start crond if needed
/app/scripts/start-crond.sh

# Start the API
exec /app/dehydrated-api-go -config ${CONFIG_FILE}