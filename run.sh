#!/bin/bash
echo "Regenerating Swagger docs..."
~/go/bin/swag init -g cmd/server/main.go -o docs

echo "Starting the server..."
go run ./cmd/server/main.go
