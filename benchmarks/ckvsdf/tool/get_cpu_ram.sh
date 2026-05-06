#!/bin/bash

get_pids() {
    process_name="$1"
    pgrep -f "$process_name" 2>/dev/null
    return $?
}

get_group_stats() {
    local label=$1
    shift
    local pids=$@

    [[ -z "$pids" ]] && {
        echo "$label: 0 0 0 0"
        return
    }

    ps -o %cpu,%mem,rss,vsz --no-headers -p $pids 2>/dev/null |
        awk -v label="$label" '
    {
        cpu+=$1; mem+=$2; rss+=$3; vsz+=$4
    }
    END {
        print label ":", cpu, mem, rss, vsz
    }'
}

while true; do
    FLASK=$(get_pids "python3 -m flask run --host 0.0.0.0")
    CKS=$(get_pids "cks -c")
    CKC=$(get_pids "ckc exploit run -e benchmark -n CookieService")
    CLIENTS=$(get_pids "python3 /tmp/DestructiveFarm/client/benchmark.py")

    get_group_stats "FLASK" $FLASK
    get_group_stats "CKS" $CKS
    get_group_stats "CKC" $CKC
    get_group_stats "CLIENTS" $CLIENTS

    sleep 0.5
done | tee ../output/stats_samples.txt
