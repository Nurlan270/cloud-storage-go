#!/bin/sh

set -e

minio server --console-address ":9001" /data &
MINIO_PID=$!

until mc alias set local http://minio:9000 "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD" >/dev/null 2>&1; do
    sleep 1
done

mc mb --ignore-existing local/user-files

wait "$MINIO_PID"