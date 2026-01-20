#!/bin/bash
# Load API key from .env and run tests
set -a
source .env
set +a

echo "Testing with MOUSER_API_KEY set"
echo "================================"
go test -v -timeout 120s ./...
