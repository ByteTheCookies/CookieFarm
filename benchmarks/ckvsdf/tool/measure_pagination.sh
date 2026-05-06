#!/bin/bash

# ==========================================
# CookieFarm Pagination Measurement
# ==========================================

# Default output file paths
OUTPUT_COLD_CF="${1:-../output/cf_latency_cold.json}"
OUTPUT_WARM_CF="${2:-../output/cf_latency_warm.json}"
OUTPUT_COLD_DF="${3:-../output/df_latency_cold.json}"
OUTPUT_WARM_DF="${4:-../output/df_latency_warm.json}"

echo "==> Authenticating to CookieFarm..."
CF_TOKEN=$(curl -s -i -X POST http://localhost:8080/api/v1/auth/login -d "password=password" | grep -i "set-cookie: token=" | sed -e 's/.*token=\([^;]*\).*/\1/' | tr -d '\r')

if [[ -z "$CF_TOKEN" ]]; then
    echo "Failed to get CookieFarm token"
    # exit 1
fi

echo "==> Running CookieFarm Cold Cache test..."
python3 measure_pagination.py \
    --url "http://localhost:8080/api/v1/flags/40" \
    --requests 50 \
    --mode cold \
    --cookie "token=${CF_TOKEN}" \
    --output "$OUTPUT_COLD_CF"

echo "==> Running CookieFarm Warm Cache test..."
python3 measure_pagination.py \
    --url "http://localhost:8080/api/v1/flags/40" \
    --requests 50 \
    --warmup 10 \
    --mode warm \
    --cookie "token=${CF_TOKEN}" \
    --output "$OUTPUT_WARM_CF"

# ==========================================
# DestructiveFarm Pagination Measurement
# ==========================================

# Note: DestructiveFarm API endpoint for flags might differ, e.g. /api/get_flags
# Update the URL to match DF's flag endpoint. Here using a placeholder /api/flags/40
DF_URL="http://localhost:5000/ui/show_flags"
DF_DATA="sploit=&team=&flag=&time-since=&time-until=&status=&checksystem_response=&page-number=34169"

echo "==> Running DestructiveFarm Cold Cache test..."
python3 measure_pagination.py \
    --url "$DF_URL" \
    --requests 50 \
    --mode cold \
    --method POST \
    --data "$DF_DATA" \
    --basic-auth ":password" \
    --output "$OUTPUT_COLD_DF"

echo "==> Running DestructiveFarm Warm Cache test..."
python3 measure_pagination.py \
    --url "$DF_URL" \
    --requests 50 \
    --warmup 10 \
    --mode warm \
    --method POST \
    --data "$DF_DATA" \
    --basic-auth ":password" \
    --output "$OUTPUT_WARM_DF"
