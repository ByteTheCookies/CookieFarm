#!/bin/sh

PORT="${PORT:-8080}"
DEBUG="${DEBUG:-false}"

CMD="/app/bin/cookieserver"

CMD="$CMD -p \"$PASSWORD\""
CMD="$CMD -P \"$PORT\""

if [ -n "$CONFIG_FROM_FILE" ]; then
    CMD="$CMD -c \"$CONFIG_FROM_FILE\""
fi

if [ "$DEBUG" = "true" ]; then
    CMD="$CMD -d"
fi

eval exec $CMD
