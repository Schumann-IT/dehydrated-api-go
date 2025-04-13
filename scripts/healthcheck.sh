#!/bin/sh

CRON_DIR=${CRON_DIR:-/etc/cron.d}

if [ -f $CRON_DIR/renew-certs ]; then
  if ! pgrep crond >/dev/null; then
    echo "crond is not running"
    exit 1
  fi
fi

# Check if API is responding
if ! curl -f http://localhost:${PORT}/health >/dev/null 2>&1; then
  echo "API health check failed"
  exit 1
fi

exit 0 