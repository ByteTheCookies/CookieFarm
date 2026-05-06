import { Banner } from "@cloudflare/kumo/components/banner";
import { Button } from "@cloudflare/kumo/components/button";
import { WarningCircleIcon } from "@phosphor-icons/react";
import { useRouteError } from "react-router";

function getErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    return error.message;
  }

  if (typeof error === "string") {
    return error;
  }

  return "The page failed to load.";
}

export function RouteError() {
  const error = useRouteError();

  return (
    <div className="p-6">
      <Banner
        variant="error"
        icon={<WarningCircleIcon weight="fill" />}
        title="Route Error"
        description={getErrorMessage(error)}
        action={
          <Button
            variant="secondary"
            onClick={() => {
              globalThis.location.reload();
            }}
          >
            Retry
          </Button>
        }
      />
    </div>
  );
}
