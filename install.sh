#!/usr/bin/env bash

set -euo pipefail

REPO="ByteTheCookies/CookieFarm"
INSTALL_DIR="/usr/local/bin"
TMP_DIR="$(mktemp -d)"
ASSET_NAME="ckc"
FINAL_NAME="ckc"

REQUIRED_CMDS=("curl" "jq")

echo "🔍 Checking for required tools..."
for cmd in "${REQUIRED_CMDS[@]}"; do
  if ! command -v "$cmd" &>/dev/null; then
    echo "❌ Error: '$cmd' is not installed. Please install it first."
    exit 1
  fi
done

# Fetch latest version info from GitHub
echo "🌐 Checking latest release..."
LATEST_VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | jq -r '.tag_name')
ASSET_URL=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | jq -r '.assets[].browser_download_url' | grep "$ASSET_NAME" || true)

if [[ -z "$ASSET_URL" || -z "$LATEST_VERSION" ]]; then
  echo "❌ Error: Unable to fetch latest release or asset."
  exit 1
fi

# # Check if binary already installed and version matches
# if command -v "$FINAL_NAME" &>/dev/null; then
#   INSTALLED_VERSION=$("$FINAL_NAME" --version | grep -oE '[0-9]+\.[0-9]+\.[0-9]+')
#   if [[ "$INSTALLED_VERSION" == "${LATEST_VERSION#v}" ]]; then
#     echo "✅ $FINAL_NAME is already up to date (v$INSTALLED_VERSION)."
#     exit 0
#   else
#     echo "🔁 Updating $FINAL_NAME from v$INSTALLED_VERSION to $LATEST_VERSION..."
#   fi
# else
#   echo "📦 $FINAL_NAME is not installed. Proceeding with fresh install..."
# fi

echo "⬇️ Downloading from: $ASSET_URL"
curl -L "$ASSET_URL" -o "$TMP_DIR/$ASSET_NAME"

echo "🚚 Installing to $INSTALL_DIR..."
sudo mv "$TMP_DIR/$ASSET_NAME" "$INSTALL_DIR/$FINAL_NAME"
sudo chmod +x "$INSTALL_DIR/$FINAL_NAME"

echo "🔧 Setting up configuration..."
CONFIG_DIR="$HOME/.config/cookiefarm"
mkdir -p "$CONFIG_DIR"
cp -r ./client/exploits/utils "$CONFIG_DIR" 2>/dev/null || true
echo "# Configuration directory for CookieFarm" > "$CONFIG_DIR/.readme"

echo "📄 Creating default configuration file..."
cat <<EOF > "$CONFIG_DIR/config.yml"
address: "localhost"
port: 8080
https: false
username: "cookieguest"
EOF

echo "✅ Installation complete! Run '$FINAL_NAME --help' to get started. Enjoy farming 🍪"

rm -rf "$TMP_DIR"


# bash <(curl -fsSL https://raw.githubusercontent.com/ByteTheCookies/CookieFarm/refs/heads/dev-akiidjk-cli/install.sh)
