#!/bin/bash

# ==========================================
# start_benchmark.sh
# Setup CookieFarm and DestructiveFarm
# and run the dummy exploits
# ==========================================

# Track background process PIDs
PIDS=()
BENCHMARK_DUR=5

# Cleanup function to kill all background processes and avoid zombie processes
cleanup() {
    echo ""
    echo "==> Stopping all background processes..."
    for PID in "${PIDS[@]}"; do
        if kill -0 "$PID" 2>/dev/null; then
            kill "$PID" 2>/dev/null
            wait "$PID" 2>/dev/null
        fi
    done
    pkill -f "/tmp/DestructiveFarm"
    pkill -f "flask run --host 0.0.0.0"
    pkill -f "get_cpu_ram.sh"
    echo "==> All processes stopped."
    return 0
}

# Trap SIGINT, SIGTERM, and EXIT to ensure cleanup is always called
trap cleanup SIGINT SIGTERM EXIT

echo "==> Setting up CookieFarm"
just server-run-prod &
PIDS+=($!)
sleep 3

echo "==> Setting up DestructiveFarm"
(cd /tmp/DestructiveFarm && ./server/start_server.sh) &
PIDS+=($!)
sleep 5

echo "Servers should be running in the background."

# echo "==> Press Enter when you are ready to START the exploits..."
# read -p "Press [Enter] key to start..."

echo "==> Starting CookieFarm Exploit..."
# Example command to run CF exploit:
just client-build-linux-prod
../../../bin/ckc login -P password
../../../bin/ckc exploit run -e benchmark -n CookieService -t 5 -W 10 &
PIDS+=($!)

echo "==> Starting DestructiveFarm Exploit..."
# Example command to run DF exploit:
(cd /tmp/DestructiveFarm/client && python3 ./start_sploit.py benchmark.py -u http://localhost:5000 --attack-period 5) &
PIDS+=($!)

echo "Exploits have been started. Monitoring..."
# We can tail logs or perform benchmark measurement steps

# Start monitoring cpu and ram usage
echo "==> Starting monitoring..."
./get_cpu_ram.sh &
CPU_RAM_PID=$!
PIDS+=($CPU_RAM_PID)

echo "==> Starting flag count timeline generation..."
./generate_flag_timeline.sh $BENCHMARK_DUR &
TIMELINE_PID=$!
PIDS+=($TIMELINE_PID)

echo "==> Benchmark is running. Waiting for $BENCHMARK_DUR seconds for completion..."
wait $TIMELINE_PID

echo "==> Benchmark duration completed. Stopping exploits and monitoring..."
pkill -f "../../../bin/ckc exploit run -e benchmark -n CookieService -t 5 -W 10"
pkill -f "python3 ./start_sploit.py benchmark.py -u http://localhost:5000 --attack-period 5"
pkill -f "python3 /tmp/DestructiveFarm/client/benchmark.py"

echo "==> Benchmark measurement finished. Stopping exploits and servers..."

echo "==> Running pagination benchmarks..."
./measure_pagination.sh

echo "==> Parsing CPU and RAM stats..."
./parse_cpu_ram.sh

./break_the_query.sh

echo "==> Generating charts..."
python3 generate_charts.py \
    --stats ../output/stats_samples.txt \
    --df-flags ../output/df_flag_count_timeline.txt \
    --cf-flags ../output/cf_flag_count_timeline.txt \
    --df-lat ../output/df_latency_cold.json \
    --cf-lat ../output/cf_latency_cold.json \
    --df-lat-warm ../output/df_latency_warm.json \
    --cf-lat-warm ../output/cf_latency_warm.json \
    --df-lat-break ../output/df_latency_cold_breakQUERY.json \
    --cf-lat-break ../output/cf_latency_cold_breakQUERY.json \
    --df-lat-warm-break ../output/df_latency_warm_breakQUERY.json \
    --cf-lat-warm-break ../output/cf_latency_warm_breakQUERY.json \
    --output ../output/charts/

echo "==> Benchmark complete! Results and charts are in the 'benchmarks/ckvsdf/output' folder."

cleanup
