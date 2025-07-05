#!/bin/sh

PORT="${PORT:-8080}"
DEBUG="${DEBUG:-false}"

if ! echo "$PORT" | grep -qE '^[0-9]+$'; then
    echo "Error: PORT must be a numeric value."
    exit 1
fi

if [ "$DEBUG" != "true" ] && [ "$DEBUG" != "false" ]; then
    echo "Error: DEBUG must be either 'true' or 'false'."
    exit 1
fi

CMD="/app/bin/cks"

ARGS="-P \"$PASSWORD\""
ARGS="$ARGS -p \"$PORT\""

if [ -n "$CONFIG_FILE" ]; then
    ARGS="$ARGS -c"
fi

if [ "$DEBUG" = "true" ]; then
    ARGS="$ARGS -D"
fi

CMD="$CMD $ARGS"
eval exec $CMD
