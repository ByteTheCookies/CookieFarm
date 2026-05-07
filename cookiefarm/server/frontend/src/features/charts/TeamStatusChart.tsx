import { useMemo } from "react";
import { Chart, ChartPalette } from "@cloudflare/kumo/components/chart";
import type { StatsSummary } from "@/api/stats";
import { ChartCard } from "./ChartCard";
import { echarts } from "./chartRuntime";

const isDarkMode = true;

export function TeamStatusChart(props: Readonly<{ summary: StatsSummary }>) {
  const stats = props.summary.flags_stats ?? [];

  const flagDistributionOption = useMemo(() => {
    return {
      tooltip: {
        trigger: "axis" as const,
        axisPointer: { type: "shadow" as const },
      },
      grid: { top: 30, right: 20, bottom: 30, left: 60 },
      xAxis: {
        type: "category" as const,
        data: stats.map((item) => `Team ${item.team_id}`),
        axisLabel: { color: "#A1A1AA" },
        axisLine: { lineStyle: { color: "#3F3F46" } },
      },
      yAxis: {
        type: "value" as const,
        axisLabel: { color: "#A1A1AA" },
        splitLine: { lineStyle: { color: "#3F3F46" } },
      },
      series: [
        {
          name: "Accepted",
          type: "bar" as const,
          stack: "total",
          data: stats.map((item) => item.accepted_flags),
          itemStyle: { color: ChartPalette.categorical(0, isDarkMode) },
        },
        {
          name: "Denied",
          type: "bar" as const,
          stack: "total",
          data: stats.map((item) => item.denied_flags),
          itemStyle: { color: ChartPalette.semantic("Attention", isDarkMode) },
        },
        {
          name: "Error",
          type: "bar" as const,
          stack: "total",
          data: stats.map((item) => item.error_flags),
          itemStyle: { color: ChartPalette.semantic("Warning", isDarkMode) },
        },
        {
          name: "Unsubmitted",
          type: "bar" as const,
          stack: "total",
          data: stats.map((item) => item.unsubmitted_flags),
          itemStyle: { color: ChartPalette.semantic("Neutral", isDarkMode) },
        },
      ],
    };
  }, [stats]);

  return (
    <ChartCard
      title="Flag Status Distribution"
      description="Team-by-team accepted, denied, error, and unsubmitted totals."
    >
      <Chart
        echarts={echarts}
        options={flagDistributionOption}
        isDarkMode={isDarkMode}
        height={320}
      />
    </ChartCard>
  );
}
