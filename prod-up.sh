#!/bin/bash
case "$1" in
  "up")
    docker-compose -f docker-compose.prod.yml up -d --remove-orphans
    ;;
  "down")
    docker-compose -f docker-compose.prod.yml down
    ;;
  "logs")
    docker-compose -f docker-compose.prod.yml logs -f
    ;;
  *)
    echo "Usage: $0 {up|down|logs}"
    exit 1
    ;;
esacъ
