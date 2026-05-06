import { cn } from "@cloudflare/kumo/utils";

type LiveDotStatus = "connecting" | "open" | "closed" | "error";

const statusClasses: Record<LiveDotStatus, string> = {
  connecting: "bg-kumo-warning",
  open: "bg-kumo-success",
  closed: "bg-kumo-subtle",
  error: "bg-kumo-danger",
};

const labels: Readonly<Record<LiveDotStatus, string>> = {
  connecting: "Connecting",
  open: "Live",
  closed: "Offline",
  error: "Error",
};

export function LiveDot(props: {
  status: LiveDotStatus;
  showLabel?: boolean;
}) {
  return (
    <div className="inline-flex items-center gap-2 text-sm text-kumo-subtle">
      <span
        aria-hidden="true"
        className={cn(
          "live-dot-pulse size-2.5 rounded-full ring-2 ring-kumo-base",
          statusClasses[props.status],
        )}
      />
      {props.showLabel ? <span>{labels[props.status]}</span> : null}
    </div>
  );
}
