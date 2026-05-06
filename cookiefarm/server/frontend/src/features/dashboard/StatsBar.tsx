import { useMemo } from "react";
import * as echarts from "echarts/core";
import { BarChart, LineChart } from "echarts/charts";
import {
  AriaComponent,
  AxisPointerComponent,
  GridComponent,
  LegendComponent,
  TooltipComponent,
} from "echarts/components";
import { CanvasRenderer } from "echarts/renderers";
import { Chart, ChartLegend, ChartPalette } from "@cloudflare/kumo/components/chart";
import type { StatsSummary } from "@/api/stats";
import { formatTickLabel } from "@/features/charts/chartData";
import type { TickPoint } from "@/features/charts/chartData";

echarts.use([
  BarChart,
  LineChart,
  AxisPointerComponent,
  AriaComponent,
  TooltipComponent,
  GridComponent,
  LegendComponent,
  CanvasRenderer,
]);

export function StatsBar(props: Readonly<{
  summary: StatsSummary;
  tickSeries: TickPoint[];
}>) {
  const isDarkMode = true;
  const stats = props.summary.flags_stats ?? [];
  const tickSeries = props.tickSeries;

  const flagsOverTimeOption = useMemo(() => {
    return {
      tooltip: {
        trigger: "axis" as const,
        axisPointer: { type: "line" as const },
      },
      grid: { top: 30, right: 20, bottom: 45, left: 55 },
      xAxis: {
        type: "category" as const,
        boundaryGap: false,
        data: tickSeries.map((point) => formatTickLabel(point.timestamp)),
        axisLabel: { color: isDarkMode ? "#A1A1AA" : "#52525B", hideOverlap: true },
        axisLine: { lineStyle: { color: isDarkMode ? "#3F3F46" : "#E4E4E7" } },
      },
      yAxis: {
        type: "value" as const,
        minInterval: 1,
        axisLabel: { color: isDarkMode ? "#A1A1AA" : "#52525B" },
        splitLine: { lineStyle: { color: isDarkMode ? "#3F3F46" : "#E4E4E7" } },
      },
      series: [
        {
          name: "Flags",
          data: tickSeries.map((point) => point.total),
          type: "line" as const,
          smooth: true,
          symbol: "circle" as const,
          symbolSize: 6,
          lineStyle: { width: 3, color: ChartPalette.semantic("Neutral", isDarkMode) },
          itemStyle: { color: ChartPalette.semantic("Neutral", isDarkMode) },
          areaStyle: { color: "rgba(173, 173, 184, 0.16)" },
        },
      ],
    };
  }, [tickSeries, isDarkMode]);

  const statusOption = useMemo(() => {
    const categories = stats.map((s) => `Team ${s.team_id}`);

    return {
      tooltip: {
        trigger: "axis" as const,
        axisPointer: { type: "shadow" as const },
      },
      grid: { top: 30, right: 20, bottom: 30, left: 60 },
      xAxis: {
        type: "category" as const,
        data: categories,
        axisLabel: { color: isDarkMode ? "#A1A1AA" : "#52525B" },
        axisLine: { lineStyle: { color: isDarkMode ? "#3F3F46" : "#E4E4E7" } },
      },
      yAxis: {
        type: "value" as const,
        axisLabel: { color: isDarkMode ? "#A1A1AA" : "#52525B" },
        splitLine: { lineStyle: { color: isDarkMode ? "#3F3F46" : "#E4E4E7" } },
      },
      series: [
        {
          name: "Accepted",
          type: "bar" as const,
          stack: "total",
          data: stats.map((s) => s.accepted_flags),
          itemStyle: { color: ChartPalette.categorical(0, isDarkMode) },
        },
        {
          name: "Denied",
          type: "bar" as const,
          stack: "total",
          data: stats.map((s) => s.denied_flags),
          itemStyle: { color: ChartPalette.semantic("Attention", isDarkMode) },
        },
        {
          name: "Error",
          type: "bar" as const,
          stack: "total",
          data: stats.map((s) => s.error_flags),
          itemStyle: { color: ChartPalette.semantic("Warning", isDarkMode) },
        },
        {
          name: "Unsubmitted",
          type: "bar" as const,
          stack: "total",
          data: stats.map((s) => s.unsubmitted_flags),
          itemStyle: { color: ChartPalette.semantic("Neutral", isDarkMode) },
        },
      ],
    };
  }, [stats, isDarkMode]);

  const totalTickFlags = tickSeries.reduce((acc, point) => acc + point.total, 0);
  const busiestTick = tickSeries.reduce(
    (current, point) => (point.total > current.total ? point : current),
    tickSeries[0] ?? { timestamp: 0, total: 0 },
  );
  const totalAcceptedCount = stats.reduce((acc, s) => acc + s.accepted_flags, 0);
  const totalDeniedCount = stats.reduce((acc, s) => acc + s.denied_flags, 0);
  const totalErrorCount = stats.reduce((acc, s) => acc + s.error_flags, 0);
  const totalUnsubmittedCount = stats.reduce((acc, s) => acc + s.unsubmitted_flags, 0);

  return (
    <div className="grid gap-4 lg:grid-cols-2">
      <div className="flex flex-col rounded-2xl border border-kumo-line bg-kumo-base p-4">
        <h3 className="mb-4 text-sm font-medium text-kumo-fg-primary">Flags Over Time</h3>
        <div className="flex-1">
          <Chart
            echarts={echarts}
            options={flagsOverTimeOption}
            isDarkMode={isDarkMode}
            height={250}
          />
        </div>
        <div className="mt-4 flex flex-wrap gap-4">
          <ChartLegend.SmallItem
            name="History"
            color={ChartPalette.semantic("Neutral", isDarkMode)}
            value={String(totalTickFlags)}
            unit="flags"
          />
          <ChartLegend.SmallItem
            name="Peak Tick"
            color={ChartPalette.categorical(1, isDarkMode)}
            value={String(busiestTick.total)}
            unit="flags"
          />
        </div>
      </div>

      <div className="flex flex-col rounded-2xl border border-kumo-line bg-kumo-base p-4">
        <h3 className="mb-4 text-sm font-medium text-kumo-fg-primary">Flag Status Distribution</h3>
        <div className="flex-1">
          <Chart
            echarts={echarts}
            options={statusOption}
            isDarkMode={isDarkMode}
            height={250}
          />
        </div>
        <div className="mt-4 flex flex-wrap gap-4">
          <ChartLegend.SmallItem
            name="Accepted"
            color={ChartPalette.categorical(0, isDarkMode)}
            value={String(totalAcceptedCount)}
          />
          <ChartLegend.SmallItem
            name="Denied"
            color={ChartPalette.semantic("Attention", isDarkMode)}
            value={String(totalDeniedCount)}
          />
          <ChartLegend.SmallItem
            name="Error"
            color={ChartPalette.semantic("Warning", isDarkMode)}
            value={String(totalErrorCount)}
          />
          <ChartLegend.SmallItem
            name="Unsubmitted"
            color={ChartPalette.semantic("Neutral", isDarkMode)}
            value={String(totalUnsubmittedCount)}
          />
        </div>
      </div>
    </div>
  );
}
