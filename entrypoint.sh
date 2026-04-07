#!/bin/sh
set -e

echo "Running migrations..."
./archipulse migrate

echo "Seeding demo data..."
./archipulse seed

echo "Starting server..."
exec ./archipulse serve
