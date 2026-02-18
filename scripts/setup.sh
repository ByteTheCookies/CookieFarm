#!/bin/bash

set -e

# === CONFIG ===
VENV_ACTIVATE=".venv/bin/activate"
FLAGCHECKER_SCRIPT="tests/flagchecker.py"
SCRIPTS_DIR="scripts"
TESTS_DIR="tests"
REQUIREMENTS="requirements.txt"

# === USAGE CHECK ===
if [[ $# -ne 2 ]]; then
    echo -e "Usage:\n  ./setup.sh <num_containers> <production_mode>\n"
    echo "  num_containers: Number of containers to start (1-10)"
    echo "  production_mode: 0 for development, 1 for production"
    exit 1
fi

# === CLEANUP HANDLER ===
cleanup() {
    echo "🧹 Cleaning up... Closing terminals and Docker..."
    kitty @ close-window --match title:flagchecker || true
    kitty @ close-window --match title:cks || true
    kitty @ close-window --match title:service || true
    kitty @ close-window --match title:frontend || true
    docker compose down
    exit
}
trap cleanup SIGINT

cd ..

# === REQUIREMENTS ===
echo "📦 Installing Python dependencies..."
pip install --upgrade pip > /dev/null
pip install -r "$REQUIREMENTS" > /dev/null

# === FLAGCHECKER ===
echo "🚩 Starting Flagchecker..."
chmod +x "$FLAGCHECKER_SCRIPT"
kitty --title "flagchecker" bash -c "source $VENV_ACTIVATE && $FLAGCHECKER_SCRIPT $1; exec bash" &
echo "✅ Flagchecker launched in a separate terminal! 🎉"

# === SERVER ===
echo "🍪 Starting CookieFarm Server..."
kitty --title "cookieserver" bash -c "just server-build-plugins; just server-run; exec bash" &
echo "✅ Server started!"

# === SERVICES ===
echo "🚀 Starting services..."
cd "$TESTS_DIR"
chmod +x ./start_containers.sh
kitty --title "service" bash -c "./start_containers.sh $1; exec bash" &
echo "✅ Services started!"

# === COMPLETION ===
echo -e "\n🎯 Cookie Farm Server ready to use!"

read -p "🔻 Press ENTER to close all terminals started by this script..."
cleanup
