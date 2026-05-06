import type { Flag } from "@/api/flags";

export type TickPoint = {
  timestamp: number;
  total: number;
  queued: number;
  accepted: number;
  denied: number;
  error: number;
  invalid: number;
};

export type ExploitShare = {
  name: string;
  value: number;
  percentage: number;
};

type TickAccumulator = Omit<TickPoint, "timestamp">;

function createEmptyAccumulator(): TickAccumulator {
  return {
    total: 0,
    queued: 0,
    accepted: 0,
    denied: 0,
    error: 0,
    invalid: 0,
  };
}

function normalizeTickSeconds(tickSeconds: number): number {
  return Number.isFinite(tickSeconds) && tickSeconds > 0 ? tickSeconds : 60;
}

function getExploitDisplayName(exploitName: string): string {
  const trimmed = exploitName.trim();
  if (!trimmed) {
    return "unknown";
  }

  const parts = trimmed.split(/[/\\]+/).filter(Boolean);
  return parts.at(-1) ?? "unknown";
}

export function buildTickSeries(flags: Flag[], tickSeconds: number): TickPoint[] {
  if (flags.length === 0) {
    return [];
  }

  const safeTickSeconds = normalizeTickSeconds(tickSeconds);
  const timestamps = flags
    .map((flag) => flag.submit_time)
    .filter((timestamp) => timestamp > 0)
    .sort((left, right) => left - right);

  if (timestamps.length === 0) {
    return [];
  }

  const firstTimestamp = timestamps[0];
  const lastTimestamp = timestamps.at(-1);
  if (firstTimestamp === undefined || lastTimestamp === undefined) {
    return [];
  }

  const minTick = Math.floor(firstTimestamp / safeTickSeconds) * safeTickSeconds;
  const maxTick = Math.floor(lastTimestamp / safeTickSeconds) * safeTickSeconds;
  const buckets = new Map<number, TickAccumulator>();

  for (let tick = minTick; tick <= maxTick; tick += safeTickSeconds) {
    buckets.set(tick, createEmptyAccumulator());
  }

  for (const flag of flags) {
    if (flag.submit_time <= 0) {
      continue;
    }

    const tick = Math.floor(flag.submit_time / safeTickSeconds) * safeTickSeconds;
    const bucket = buckets.get(tick) ?? createEmptyAccumulator();
    bucket.total += 1;

    switch (flag.status) {
      case 0:
        bucket.queued += 1;
        break;
      case 1:
        bucket.accepted += 1;
        break;
      case 2:
        bucket.denied += 1;
        break;
      case 3:
        bucket.error += 1;
        break;
      case 4:
        bucket.invalid += 1;
        break;
      default:
        break;
    }

    buckets.set(tick, bucket);
  }

  return Array.from(buckets.entries())
    .sort(([left], [right]) => left - right)
    .map(([timestamp, counts]) => ({
      timestamp,
      ...counts,
    }));
}

export function buildExploitShare(flags: Flag[]): ExploitShare[] {
  if (flags.length === 0) {
    return [];
  }

  const counts = new Map<string, number>();

  for (const flag of flags) {
    const name = getExploitDisplayName(flag.exploit_name);
    counts.set(name, (counts.get(name) ?? 0) + 1);
  }

  const total = flags.length;

  return Array.from(counts.entries())
    .sort((left, right) => right[1] - left[1] || left[0].localeCompare(right[0]))
    .map(([name, value]) => ({
      name,
      value,
      percentage: total === 0 ? 0 : (value / total) * 100,
    }));
}

export function formatTickLabel(timestamp: number): string {
  return new Date(timestamp * 1000).toLocaleString([], {
    month: "short",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}
