#!/bin/sh

# the dehydrated base dir, variable for testing
BASE_DIR=${BASE_DIR:-/data/dehydrated}
# the cron dir, variable for testing
CRON_DIR=${CRON_DIR:-/etc/cron.d}
# the app user, variable for testing
APP_USER=${APP_USER:-root}

# no schedule: no need to configure cron
if [ -z "$CRON_SCHEDULE" ]; then
  exit 0
fi

# cron dir
mkdir -p ${CRON_DIR}

# Create the cron job
echo "${CRON_SCHEDULE} /app/scripts/renew-certs.sh" > ${CRON_DIR}/renew-certs
