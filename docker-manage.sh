#!/bin/bash

# A simple script to manage the Docker container

COMMAND=$1

case $COMMAND in
  build)
    echo "Building Docker image..."
    docker compose build
    ;;
  up)
    echo "Starting container..."
    docker compose up -d
    ;;
  down)
    echo "Stopping and removing container..."
    docker compose down
    ;;
  logs)
    docker compose logs -f
    ;;
  restart)
    echo "Restarting container..."
    docker compose restart
    ;;
  status)
    docker compose ps
    ;;
  *)
    echo "Usage: $0 {build|up|down|logs|restart|status}"
    exit 1
    ;;
esac
