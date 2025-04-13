#!/bin/sh

CRON_DIR=${CRON_DIR:-/etc/cron.d}

if [ ! -f $CRON_DIR/renew-certs ]; then
  exit 0
fi

# start cron
crond -f -l 8 &