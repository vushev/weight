#!/bin/bash
case "$1" in
  "up")
    docker-compose -f docker-compose.dev.yml up -d
    ;;
  "down")
    docker-compose -f docker-compose.dev.yml down
    ;;
  "logs")
    docker-compose -f docker-compose.dev.yml logs -f
    ;;
  *)
    echo "Usage: $0 {up|down|logs}"
    exit 1
    ;;
esac