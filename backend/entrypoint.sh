#!/bin/sh
# Fix ownership for mounted volumes (runs as root, then drops to confbox user)
chown -R confbox:confbox /data 2>/dev/null || true
exec su-exec confbox /confbox "$@"
