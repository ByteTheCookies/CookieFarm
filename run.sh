#!/bin/sh

PORT="${PORT:-8080}"
DEBUG="${DEBUG:-false}"

CMD="/app/bin/cks"

CMD="$CMD -P \"$PASSWORD\""
CMD="$CMD -p \"$PORT\""pacman -S docker-buildx

if [ -n "$CONFIG_FILE" ]; then
    CMD="$CMD -c"
fi

if [ "$DEBUG" = "true" ]; then
    CMD="$CMD -D"
fi

eval exec $CMD
