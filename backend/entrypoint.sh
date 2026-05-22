#!/bin/sh
# Fix ownership for mounted volumes (runs as root, then drops to configbox user)
chown -R configbox:configbox /data 2>/dev/null || true
exec su-exec configbox /configbox "$@"
