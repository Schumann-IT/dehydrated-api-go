#!/bin/sh

export CONFIG_FILE=/app/config/config.yaml

# update api config from env variables
/app/scripts/update-config.sh

# Start the API
exec /app/dehydrated-api-go -config ${CONFIG_FILE}