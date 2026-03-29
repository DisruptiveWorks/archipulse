#!/bin/sh
set -e

echo "Running migrations..."
./archipulse migrate

echo "Starting server..."
exec ./archipulse serve
