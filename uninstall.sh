#!/usr/bin/env bash

set -euo pipefail

BINARY_PATH="/usr/local/bin/ckc"
CONFIG_PATH="$HOME/.config/cookiefarm"

echo "🧹 Starting CookieFarm uninstallation..."

# Remove binary
if [[ -f "$BINARY_PATH" ]]; then
  echo "❌ Removing binary from $BINARY_PATH..."
  sudo rm "$BINARY_PATH"
else
  echo "ℹ️ No binary found at $BINARY_PATH. Skipping."
fi

# Remove config
if [[ -d "$CONFIG_PATH" ]]; then
  echo "🗑️ Deleting configuration at $CONFIG_PATH..."
  rm -rf "$CONFIG_PATH"
else
  echo "ℹ️ No configuration found at $CONFIG_PATH. Skipping."
fi

echo "✅ CookieFarm has been successfully uninstalled. See you next time! 🍪"
