import { useState } from "react";
import { Banner } from "@cloudflare/kumo/components/banner";
import { Breadcrumbs } from "@cloudflare/kumo/components/breadcrumbs";
import { Button } from "@cloudflare/kumo/components/button";
import { Input } from "@cloudflare/kumo/components/input";
import { useKumoToastManager } from "@cloudflare/kumo/components/toast";
import { ArrowSquareOutIcon, WarningCircleIcon } from "@phosphor-icons/react";
import { Link } from "react-router";
import { useConfig } from "@/api/config";
import { deleteFlag, submitFlag, useFlags } from "@/api/flags";
import { useChartStats, useStatsSummary } from "@/api/stats";
import { FlagTable } from "@/features/flags/FlagTable";
import { PageHeader } from "@/components/kumo/page-header/page-header";
import { StatsBar } from "./StatsBar";

const dashboardRefreshInterval = 10_000;

const breadcrumbs = (
  <Breadcrumbs className="px-3 py-2 text-sm">
    <Breadcrumbs.Link href="/">Operations</Breadcrumbs.Link>
    <Breadcrumbs.Separator />
    <Breadcrumbs.Current>Dashboard</Breadcrumbs.Current>
  </Breadcrumbs>
);

export function DashboardPage() {
  const toast = useKumoToastManager();
  const config = useConfig();
  const summaryQuery = useStatsSummary({ refreshInterval: dashboardRefreshInterval });
  const chartStatsQuery = useChartStats(config.server.tick_time, {
    refreshInterval: dashboardRefreshInterval,
  });
  const flagsQuery = useFlags(
    { limit: 25 },
    { refreshInterval: dashboardRefreshInterval },
  );
  const summary = summaryQuery.data!;
  const chartStats = chartStatsQuery.data!;
  const flags = flagsQuery.data!.flags;
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [flagCode, setFlagCode] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [deleting, setDeleting] = useState(false);

  async function refreshDashboard(): Promise<void> {
    await Promise.all([
      summaryQuery.mutate(),
      chartStatsQuery.mutate(),
      flagsQuery.mutate(),
    ]);
    setErrorMessage(null);
  }

  const swrError = summaryQuery.error ?? chartStatsQuery.error ?? flagsQuery.error;
  const visibleErrorMessage =
    errorMessage ?? (swrError instanceof Error ? swrError.message : null);

  return (
    <div className="space-y-6">
      <PageHeader
        breadcrumbs={breadcrumbs}
        title="Dashboard"
        description="Collector overview, manual submit/delete actions, and recent stored flags."
      >
        <Link
          to="/flags"
          className="inline-flex h-9 items-center gap-2 rounded-lg bg-kumo-brand px-3 text-base text-white hover:bg-kumo-brand-hover"
        >
          Open Flag Feed
          <ArrowSquareOutIcon size={16} />
        </Link>
      </PageHeader>

      {visibleErrorMessage ? (
        <Banner
          variant="error"
          icon={<WarningCircleIcon weight="fill" />}
          title="Refresh failed"
          description={visibleErrorMessage}
        />
      ) : null}

      {config.configured == false ? (
        <Banner
          variant="error"
          title="Configuration incomplete"
          description="The backend config is not marked as configured yet."
        />
      ) : null}

      <StatsBar
        summary={summary}
        tickSeries={chartStats.tick_series}
      />

      <section className="rounded-2xl border border-kumo-line bg-kumo-base p-4">
        <div className="flex flex-col gap-4 xl:flex-row xl:items-end xl:justify-between">
          <div className="grid gap-4 md:grid-cols-[minmax(0,1fr)_auto_auto] items-end">
            <Input
              label="Manual Flag"
              placeholder="FLAG{...}"
              value={flagCode}
              onChange={(event) => {
                setFlagCode(event.target.value);
              }}
            />
            <Button
              loading={submitting}
              onClick={() => {
                const code = flagCode.trim();
                if (!code) {
                  return;
                }

                setSubmitting(true);
                void submitFlag({
                  flag_code: code,
                  service_name: "unknown",
                  port_service: 0,
                  submit_time: Math.floor(Date.now() / 1000),
                  response_time: 0,
                  msg: "Manual submission from dashboard",
                  status: 0,
                  team_id: 0,
                  username: "dashboard",
                  exploit_name: "manual",
                })
                  .then(async () => {
                    setFlagCode("");
                    toast.add({
                      variant: "success",
                      title: "Flag submitted",
                      description: "The flag has been queued for processing.",
                    });
                    await refreshDashboard();
                  })
                  .catch((error: unknown) => {
                    toast.add({
                      variant: "error",
                      title: "Submit failed",
                      description:
                        error instanceof Error ? error.message : "Unable to submit the flag.",
                    });
                  })
                  .finally(() => {
                    setSubmitting(false);
                  });
              }}
            >
              Submit
            </Button>
            <Button
              variant="secondary"
              loading={deleting}
              onClick={() => {
                const code = flagCode.trim();
                if (!code) {
                  return;
                }

                setDeleting(true);
                void deleteFlag(code)
                  .then(async () => {
                    setFlagCode("");
                    toast.add({
                      variant: "success",
                      title: "Flag deleted",
                      description: "The matching flag row has been removed.",
                    });
                    await refreshDashboard();
                  })
                  .catch((error: unknown) => {
                    toast.add({
                      variant: "error",
                      title: "Delete failed",
                      description:
                        error instanceof Error ? error.message : "Unable to delete the flag.",
                    });
                  })
                  .finally(() => {
                    setDeleting(false);
                  });
              }}
            >
              Delete
            </Button>
          </div>

          <div className="flex flex-wrap items-center gap-3">
            <Button
              variant="secondary"
              onClick={() => {
                void refreshDashboard().catch((error: unknown) => {
                  setErrorMessage(
                    error instanceof Error ? error.message : "Dashboard refresh failed",
                  );
                });
              }}
            >
              Reload
            </Button>
          </div>
        </div>
      </section>

      <section className="space-y-3">
        <div>
          <h2 className="text-lg font-semibold text-kumo-fg-primary">Latest Flags</h2>
          <p className="text-sm text-kumo-fg-secondary mt-1">
            The newest 25 rows. Dashboard charts use aggregated stats, not the full flag archive.
          </p>
        </div>
        <FlagTable
          rows={flags}
          emptyTitle="No flags stored yet"
          emptyDescription="Submit a manual flag or wait for the collector to flush new rows."
        />
      </section>
    </div>
  );
}
