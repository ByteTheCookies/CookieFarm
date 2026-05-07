import { ChartLegend, ChartPalette } from "@cloudflare/kumo/components/chart";
import type { ChartStats } from "@/api/stats";

const isDarkMode = true;

export function ChartsSummary(props: Readonly<{ chartStats: ChartStats }>) {
  const tickSeries = props.chartStats.tick_series;
  const exploitShare = props.chartStats.exploit_share;
  const latestTickCount = tickSeries[tickSeries.length - 1]?.total ?? 0;
  const leadingExploit = exploitShare[0];

  return (
    <section className="flex flex-wrap gap-4 rounded-2xl border border-kumo-line bg-kumo-base p-4">
      <ChartLegend.LargeItem
        name="History"
        color={ChartPalette.semantic("Neutral", isDarkMode)}
        value={String(props.chartStats.total_flags)}
        unit="flags"
      />
      <ChartLegend.LargeItem
        name="Latest Tick"
        color={ChartPalette.categorical(0, isDarkMode)}
        value={String(latestTickCount)}
        unit="flags"
      />
      {leadingExploit ? (
        <ChartLegend.LargeItem
          name="Top Exploit"
          color={ChartPalette.categorical(1, isDarkMode)}
          value={`${leadingExploit.percentage.toFixed(1)}%`}
          unit={leadingExploit.name}
        />
      ) : (
        <ChartLegend.LargeItem
          name="Top Exploit"
          color={ChartPalette.categorical(1, isDarkMode)}
          value="0%"
        />
      )}
    </section>
  );
}
