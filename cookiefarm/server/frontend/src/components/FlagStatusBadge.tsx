import { Badge } from "@cloudflare/kumo/components/badge";
import type { BadgeVariant } from "@cloudflare/kumo/components/badge";
import type { FlagStatus } from "@/api/flags";

const variants = {
  0: { variant: "neutral", label: "Queued" },
  1: { variant: "success", label: "Accepted" },
  2: { variant: "warning", label: "Denied" },
  3: { variant: "error", label: "Error" },
  4: { variant: "warning", label: "Invalid" },
} satisfies Record<FlagStatus, { label: string; variant: BadgeVariant }>;

export function FlagStatusBadge(props: Readonly<{ status: FlagStatus }>) {
  const config = variants[props.status];
  return (
    <Badge variant={config.variant} className="text-[11px]">
      {config.label}
    </Badge>
  );
}
