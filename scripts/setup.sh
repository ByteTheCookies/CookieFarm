#!/bin/bash

if [[ $# -ne 1 ]]; then
	printf "Usage:\n  ./setup.sh <num_containers>"
	exit
fi

# Install requirements
pip install --upgrade pip > /dev/null
pip install -r ../requirements.txt > /dev/null

activate_venv="source ../venv/bin/activate"

# Run Flagchecker
echo "🚩 Avvio Flagchecker..."

chmod +x ../tests/flagchecker.py
kitty --title "flagchecker" bash -c "${activate_venv} && ../tests/flagchecker.py; exec bash" &

echo "✅ Flagchecker lanciato in un terminale separato! 🎉"
echo ""

# Run Server
echo "🍪 Avvio CookieFarm Server..."

cd ../server/backend/

kitty --title "cookieserver" bash -c "make build && make run ARGS='--debug'; exec bash" &

echo "✅ Server avviato!"
sleep 2
echo "📡 Invio configurazione..."

cd ../../scripts/
chmod +x shitcurl.py
./shitcurl.py

echo "✅ Configurazione inviata!"

# Run FE
echo "🌐 Start frontend"
cd ../server/frontend/
kitty --title "frontend" bash -c "/bin/bun run dev; exec bash" &
echo "🌐 Frontend started"

# Run Services
echo "🚀 Avvio Servizi..."

cd ../../tests
chmod +x ./start_container.sh
kitty --title "service" bash -c "./start_containers.sh $1; exec bash" &

echo "🚀 Servizi avviati!"

echo "🎯 Cookie Farm Server pronto all'uso!"

# Attendi input per terminare tutti i terminali kitty
read -p "🔻 Premi INVIO per chiudere tutti i terminali avviati dallo script..."

# Chiudi le finestre kitty con i titoli assegnati
kitty @ close-window --match title:flagchecker
kitty @ close-window --match title:cookieserver
kitty @ close-window --match title:service
kitty @ close-window --match title:frontend
docker compose down

echo "🧹 Tutti i terminali sono stati chiusi!"
