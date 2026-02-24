#!/bin/bash

if [[ $# -ne 2 ]]; then
    printf "Usage:\n  ./setup.sh <num_containers> <path_df>"
    exit
fi

cleanup() {
    echo "🧹 Cleaning up... Closing terminals and Docker..."
    kitty @ close-window --match title:flagchecker
    kitty @ close-window --match title:cookieserver
    kitty @ close-window --match title:service
    kitty @ close-window --match title:frontend
    cd ../tests/
    docker compose down
    cd ../scripts/
    exit
}

trap cleanup SIGINT

# Install requirements
pip install --upgrade pip > /dev/null
pip install -r ../requirements.txt > /dev/null

activate_venv="source ../venv/bin/activate"

# Run Flagchecker
echo "🚩 Starting Flagchecker..."

chmod +x ../tests/flagchecker.py
kitty --title "flagchecker" bash -c "${activate_venv} && ../tests/flagchecker.py; exec bash" &

echo "✅ Flagchecker launched in a separate terminal! 🎉"
echo ""

# Run Services
echo "🚀 Starting Services..."

cd ../tests
chmod +x ./start_containers.sh
kitty --title "service" bash -c "./start_containers.sh $1; exec bash" &

echo "🚀 Services started!"

# Run DestructiveFarm
echo "🚀 Starting DestructiveFarm..."

cd ../scripts/
cat ./config_df.py > $2/server/config.py
chmod +x $2/server/start_server.sh
kitty --title "destructivefarm" bash -c "$2/server/start_server.sh; exec bash" &

echo "🚀 DestructiveFarm started!"

echo "🎯 DF environment is ready to use!"

# Wait for input to close all kitty terminals
read -p "🔻 Press ENTER to close all terminals started by the script..."

cleanup()

echo "🧹 All terminals have been closed!"
