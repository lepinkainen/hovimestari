#!/bin/sh
# Run all imports using project-root config, then show what would be sent to the LLM.
set -e

BINARY="./build/hovimestari"
CONFIG="--config=./config.json"

echo "=== Importing calendar ==="
$BINARY $CONFIG import-calendar

echo "=== Importing weather ==="
$BINARY $CONFIG import-weather

echo "=== Importing school lunch ==="
$BINARY $CONFIG import-school-lunch

echo "=== Importing electricity prices ==="
$BINARY $CONFIG import-electricity-price

echo "=== Brief context ==="
$BINARY $CONFIG show-brief-context

# Uncomment to generate a real brief (CLI only, no Discord/Telegram):
# $BINARY $CONFIG generate-brief --cli-only
