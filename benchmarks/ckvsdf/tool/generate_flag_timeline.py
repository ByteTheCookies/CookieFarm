#!/usr/bin/env python3
"""
generate_flag_timeline.py — Records flag count growth over time during a benchmark run.

Polls multiple SQLite databases simultaneously (same tick using threads) and writes:
    <epoch_ms> <count>
to separate output files.

Usage:
    # Multi-source synchronized polling (recommended):
    python3 generate_flag_timeline.py \
        --source "CF:../../../cookiefarm/cookiefarm.db:SELECT COUNT(*) FROM flags:../output/cf_flag_count_timeline.txt" \
        --source "DF:/tmp/DestructiveFarm/server/flags.sqlite:SELECT COUNT(*) FROM flags:../output/df_flag_count_timeline.txt" \
        --duration 600

    # Single-source legacy mode:
    python3 generate_flag_timeline.py \
        --db cookiefarm.db \
        --output ../output/cf_flag_count_timeline.txt \
        --duration 600
"""

import argparse
import os
import signal
import sqlite3
import sys
import threading
import time

# ── Args ───────────────────────────────────────────────────────────────────────

parser = argparse.ArgumentParser(
    description="Poll SQLite databases simultaneously and record flag count timelines."
)
parser.add_argument(
    "--source",
    action="append",
    default=[],
    metavar="LABEL:DB_PATH:QUERY:OUTPUT",
    help=(
        "A source definition. Format: LABEL:DB_PATH:QUERY:OUTPUT_FILE. "
        "Repeatable for multiple simultaneous sources."
    ),
)
# Legacy single-source args
parser.add_argument("--db", required=False, help="(legacy) Path to a single SQLite DB")
parser.add_argument(
    "--url", required=False, help="(legacy) URL of a flag count endpoint"
)
parser.add_argument(
    "--query",
    default="SELECT COUNT(*) FROM flags",
    help="(legacy) SQL query for single --db mode",
)
parser.add_argument(
    "--output", required=False, help="(legacy) Output file for single source"
)
parser.add_argument(
    "--field", default="count", help="(legacy) JSON field for --url mode"
)
parser.add_argument("--header", action="append", default=[], metavar="KEY:VALUE")
parser.add_argument(
    "--interval",
    type=float,
    default=0.5,
    help="Polling interval in seconds (default: 0.5)",
)
parser.add_argument(
    "--duration", type=int, default=0, help="Stop after N seconds (0 = Ctrl+C)"
)
parser.add_argument("--timeout", type=int, default=5, help="DB/HTTP timeout in seconds")
args = parser.parse_args()

# ── Build source list ──────────────────────────────────────────────────────────


class Source:
    def __init__(self, label: str, db: str, query: str, output: str):
        self.label = label
        self.db = db
        self.query = query
        self.output = output
        self.samples: list[tuple[int, int]] = []
        self.lock = threading.Lock()


sources: list[Source] = []

for raw in args.source:
    parts = raw.split(":", 3)
    if len(parts) != 4:
        print(
            f"ERROR: --source must be LABEL:DB_PATH:QUERY:OUTPUT_FILE, got: {raw}",
            file=sys.stderr,
        )
        sys.exit(1)
    label, db, query, output = parts
    sources.append(Source(label.strip(), db.strip(), query.strip(), output.strip()))

# Legacy fallback
if not sources:
    if not args.db and not args.url:
        parser.error(
            "Must provide at least one --source, or --db / --url for legacy mode."
        )
    if not args.output:
        parser.error("--output is required in legacy single-source mode.")
    label = "DB" if args.db else "URL"
    sources.append(Source(label, args.db or args.url, args.query, args.output))

# ── Helpers ────────────────────────────────────────────────────────────────────


def poll_sqlite(source: Source, timeout: int) -> int | None:
    try:
        conn = sqlite3.connect(source.db, timeout=timeout)
        cursor = conn.cursor()
        cursor.execute(source.query)
        row = cursor.fetchone()
        conn.close()
        return int(row[0]) if row is not None else None
    except Exception as e:
        print(f"\n  [{source.label}] DB Error: {e}")
        return None


def flush_source(source: Source):
    os.makedirs(
        os.path.dirname(source.output) if os.path.dirname(source.output) else ".",
        exist_ok=True,
    )
    with open(source.output, "w") as f:
        for epoch_ms, count in source.samples:
            f.write(f"{epoch_ms} {count}\n")
    print(f"  ✓ [{source.label}] Saved {len(source.samples)} samples → {source.output}")


# ── Signal handling ────────────────────────────────────────────────────────────

stop_event = threading.Event()


def handle_interrupt(sig, frame):
    print("\n==> Stopping...")
    stop_event.set()


signal.signal(signal.SIGINT, handle_interrupt)
signal.signal(signal.SIGTERM, handle_interrupt)

# ── Worker thread per source ───────────────────────────────────────────────────

# A barrier ensures all threads poll at the same tick
barrier = threading.Barrier(len(sources))


def worker(source: Source, start_time: float):
    last_count = 0
    consecutive_errors = 0
    MAX_ERRORS = 10

    while not stop_event.is_set():
        # Synchronize: all threads reach this point together before polling
        try:
            barrier.wait(timeout=args.interval + 1.0)
        except threading.BrokenBarrierError:
            break

        if stop_event.is_set():
            break

        now_ms = int(time.time() * 1000)
        elapsed = time.time() - start_time
        count = poll_sqlite(source, args.timeout)

        if count is None:
            consecutive_errors += 1
            if consecutive_errors >= MAX_ERRORS:
                print(f"\n  [{source.label}] ERROR: {MAX_ERRORS} consecutive failures.")
                stop_event.set()
        else:
            with source.lock:
                source.samples.append((now_ms, count))
            delta = count - last_count
            print(
                f"  [{elapsed:6.1f}s] {source.label}: flags={count:6d}  Δ={delta:+5d}",
                end="\r",
            )
            last_count = count
            consecutive_errors = 0

        # Check duration limit
        if args.duration > 0 and elapsed >= args.duration:
            stop_event.set()


# ── Ticker thread: controls the shared interval ───────────────────────────────


def ticker():
    """Wakes up every `interval` seconds and resets the barrier so all workers fire together."""
    while not stop_event.is_set():
        time.sleep(args.interval)
        if stop_event.is_set():
            break
        try:
            barrier.reset()
        except Exception:
            pass


# ── Main ───────────────────────────────────────────────────────────────────────

for s in sources:
    os.makedirs(
        os.path.dirname(s.output) if os.path.dirname(s.output) else ".", exist_ok=True
    )

print("==> Starting synchronized flag count timeline polling")
for s in sources:
    print(f"  [{s.label}] DB: {s.db}")
    print(f"  [{s.label}] Query: {s.query}")
    print(f"  [{s.label}] Output: {s.output}")
print(f"  Interval: {args.interval}s | Duration: {args.duration or '∞'}s")
print("Press Ctrl+C to stop early.\n")

start_time = time.time()
threads = []

for source in sources:
    t = threading.Thread(target=worker, args=(source, start_time), daemon=True)
    threads.append(t)
    t.start()

# Wait for all threads (they stop via stop_event)
for t in threads:
    t.join()

print()
for s in sources:
    flush_source(s)

if args.duration > 0:
    print(f"\nDuration limit reached ({args.duration}s).")
