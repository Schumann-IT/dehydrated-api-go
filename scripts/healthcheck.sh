#!/bin/sh

# Check if API is responding
if ! curl -f http://localhost:${PORT}/api/v1/health >/dev/null 2>&1; then
  echo "API health check failed"
  exit 1
fi

exit 0 