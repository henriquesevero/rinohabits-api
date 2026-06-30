#!/bin/sh
set -e

echo "running migrations..."
./migrate -direction up

echo "starting api..."
exec ./api
