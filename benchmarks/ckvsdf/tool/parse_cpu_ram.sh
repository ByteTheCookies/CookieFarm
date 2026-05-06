#!/bin/bash

STATS_FILE="../output/stats_samples.txt"
if [[ ! -f "$STATS_FILE" ]]; then
    echo "Error: $STATS_FILE not found." >&2
    exit 1
fi

CPU_AVG_AWK='{sum+=$2} END {if(NR>0) print sum/NR; else print 0}'

echo "Parsing $STATS_FILE..."

# Calculate averages for each group
echo -e "\n=== CPU Averages (%) ==="
echo "FLASK (DF Server): $(grep "^FLASK:" "$STATS_FILE" | awk "$CPU_AVG_AWK")"
echo "CKS (CF Server):   $(grep "^CKS:" "$STATS_FILE" | awk "$CPU_AVG_AWK")"
echo "CKC (CF Client):   $(grep "^CKC:" "$STATS_FILE" | awk "$CPU_AVG_AWK")"
echo "CLIENTS (DF):      $(grep "^CLIENTS:" "$STATS_FILE" | awk "$CPU_AVG_AWK")"

echo -e "\n=== RAM Peak (MB) ==="
echo "FLASK (DF Server): $(grep "^FLASK:" "$STATS_FILE" | awk '{if($4>max) max=$4} END {print max/1024}')"
echo "CKS (CF Server):   $(grep "^CKS:" "$STATS_FILE" | awk '{if($4>max) max=$4} END {print max/1024}')"
echo "CKC (CF Client):   $(grep "^CKC:" "$STATS_FILE" | awk '{if($4>max) max=$4} END {print max/1024}')"
echo "CLIENTS (DF):      $(grep "^CLIENTS:" "$STATS_FILE" | awk '{if($4>max) max=$4} END {print max/1024}')"

echo -e "\nParsing completed. You can run generate_charts.py using --stats $STATS_FILE"
