import { useMemo } from "react";
import { Chart, ChartLegend, ChartPalette } from "@cloudflare/kumo/components/chart";
import type { ChartEvents, KumoChartOption } from "@cloudflare/kumo/components/chart";
import type { ChartStats } from "@/api/stats";
import { ChartCard } from "./ChartCard";
import { formatTickLabel } from "./chartData";
import { echarts } from "./chartRuntime";

const isDarkMode = true;

export function TimeCharts(props: Readonly<{
  chartStats: ChartStats;
  dataZoom: NonNullable<KumoChartOption["dataZoom"]>;
  onDataZoom: ChartEvents["datazoom"];
}>) {
  const tickSeries = props.chartStats.tick_series;
  const exploitTickSeries = props.chartStats.exploit_tick_series;

  const flagsOverTimeOption = useMemo(() => {
    return {
      tooltip: {
        trigger: "axis" as const,
        axisPointer: { type: "line" as const },
      },
      animation: true,
      dataZoom: props.dataZoom,
      grid: { top: 30, right: 20, bottom: 55, left: 55 },
      xAxis: {
        type: "category" as const,
        boundaryGap: false,
        data: tickSeries.map((point) => formatTickLabel(point.timestamp)),
        axisLabel: { color: "#A1A1AA", hideOverlap: true },
        axisLine: { lineStyle: { color: "#3F3F46" } },
      },
      yAxis: {
        type: "value" as const,
        minInterval: 1,
        axisLabel: { color: "#A1A1AA" },
        splitLine: { lineStyle: { color: "#3F3F46" } },
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
  }, [props.dataZoom, tickSeries]);

  const exploitPerTickOption = useMemo(() => {
    const ticks = tickSeries.map((point) => point.timestamp);

    return {
      tooltip: {
        trigger: "axis" as const,
        axisPointer: { type: "line" as const },
      },
      animation: true,
      legend: {
        type: "scroll" as const,
        top: 0,
        textStyle: { color: "#D4D4D8" },
      },
      dataZoom: props.dataZoom,
      grid: { top: 55, right: 20, bottom: 55, left: 55 },
      xAxis: {
        type: "category" as const,
        boundaryGap: false,
        data: ticks.map((timestamp) => formatTickLabel(timestamp)),
        axisLabel: { color: "#A1A1AA", hideOverlap: true },
        axisLine: { lineStyle: { color: "#3F3F46" } },
      },
      yAxis: {
        type: "value" as const,
        minInterval: 1,
        axisLabel: { color: "#A1A1AA" },
        splitLine: { lineStyle: { color: "#3F3F46" } },
      },
      series: exploitTickSeries.map((item, index) => {
        const valuesByTick = new Map(item.data.map((point) => [point.timestamp, point.value]));

        return {
          name: item.name,
          type: "line" as const,
          smooth: true,
          symbolSize: 5,
          data: ticks.map((timestamp) => valuesByTick.get(timestamp) ?? 0),
          lineStyle: { width: 2, color: ChartPalette.categorical(index, isDarkMode) },
          itemStyle: { color: ChartPalette.categorical(index, isDarkMode) },
        };
      }),
    };
  }, [exploitTickSeries, props.dataZoom, tickSeries]);

  const statusPerTickOption = useMemo(() => {
    return {
      tooltip: {
        trigger: "axis" as const,
        axisPointer: { type: "line" as const },
      },
      animation: true,
      legend: {
        top: 0,
        textStyle: { color: "#D4D4D8" },
      },
      dataZoom: props.dataZoom,
      grid: { top: 50, right: 20, bottom: 55, left: 55 },
      xAxis: {
        type: "category" as const,
        boundaryGap: false,
        data: tickSeries.map((point) => formatTickLabel(point.timestamp)),
        axisLabel: { color: "#A1A1AA", hideOverlap: true },
        axisLine: { lineStyle: { color: "#3F3F46" } },
      },
      yAxis: {
        type: "value" as const,
        minInterval: 1,
        axisLabel: { color: "#A1A1AA" },
        splitLine: { lineStyle: { color: "#3F3F46" } },
      },
      series: [
        {
          name: "Accepted",
          type: "line" as const,
          smooth: true,
          data: tickSeries.map((point) => point.accepted),
          itemStyle: { color: ChartPalette.categorical(0, isDarkMode) },
          lineStyle: { color: ChartPalette.categorical(0, isDarkMode), width: 2 },
        },
        {
          name: "Denied",
          type: "line" as const,
          smooth: true,
          data: tickSeries.map((point) => point.denied),
          itemStyle: { color: ChartPalette.semantic("Attention", isDarkMode) },
          lineStyle: { color: ChartPalette.semantic("Attention", isDarkMode), width: 2 },
        },
        {
          name: "Error",
          type: "line" as const,
          smooth: true,
          data: tickSeries.map((point) => point.error),
          itemStyle: { color: ChartPalette.semantic("Warning", isDarkMode) },
          lineStyle: { color: ChartPalette.semantic("Warning", isDarkMode), width: 2 },
        },
        {
          name: "Queued",
          type: "line" as const,
          smooth: true,
          data: tickSeries.map((point) => point.queued),
          itemStyle: { color: ChartPalette.semantic("Neutral", isDarkMode) },
          lineStyle: { color: ChartPalette.semantic("Neutral", isDarkMode), width: 2 },
        },
      ],
    };
  }, [props.dataZoom, tickSeries]);

  const totalTickFlags = tickSeries.reduce((acc, point) => acc + point.total, 0);
  const busiestTick = tickSeries.reduce(
    (current, point) => (point.total > current.total ? point : current),
    tickSeries[0] ?? { timestamp: 0, total: 0 },
  );

  return (
    <>
      <ChartCard title="Flags Over Time" className="flex flex-col">
        <div className="flex-1">
          <Chart
            echarts={echarts}
            options={flagsOverTimeOption}
            isDarkMode={isDarkMode}
            onEvents={{ datazoom: props.onDataZoom }}
          />
        </div>
        <div className="flex flex-wrap mt-4 gap-4">
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
      </ChartCard>

      <ChartCard
        title="Flags Per Tick"
        description="Submitted flags grouped by collector tick, split into one line per exploit."
      >
        <Chart
          echarts={echarts}
          options={exploitPerTickOption}
          isDarkMode={isDarkMode}
          height={320}
          onEvents={{ datazoom: props.onDataZoom }}
        />
      </ChartCard>

      <ChartCard
        title="Status Per Tick"
        description="Accepted, denied, error, and queued trends across time."
      >
        <Chart
          echarts={echarts}
          options={statusPerTickOption}
          isDarkMode={isDarkMode}
          height={320}
          onEvents={{ datazoom: props.onDataZoom }}
        />
      </ChartCard>
    </>
  );
}
