#!/bin/bash
QUEUE=$HOME/.local/state/scratchpad-queue
touch "$QUEUE" 2>/dev/null || true

COMMAND=$1

if [ -z "$COMMAND" ]; then
  echo "Usage: $0 <push|pop>"
  exit 1
fi

if [ "$COMMAND" = "push" ]; then
  # Push IDs of moved windows
  aerospace-scratchpad move --output=json \
    | jq -r 'select(.result=="ok") | .window_id' \
    >> "$QUEUE"
  exit 0
fi

if [ "$COMMAND" = "pop" ]; then
  # Pop oldest window ID and summon it back
  if [ -s "$QUEUE" ]; then
    wid=$(head -n1 "$QUEUE")
    tail -n +2 "$QUEUE" > "$QUEUE.tmp" && mv "$QUEUE.tmp" "$QUEUE"
    aerospace-scratchpad show ".+" --filter window-id="^${wid}$" --output=json
  else
    echo "Queue empty"
  fi
fi
