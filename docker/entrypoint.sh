#!/bin/sh

mkdir -p /app/data/scripts /app/data/logs /app/data/backups

nginx

exec /app/daidai-server
