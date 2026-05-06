import { useActionState, useEffect, useRef, useState } from "react";
import { Banner } from "@cloudflare/kumo/components/banner";
import { Collapsible } from "@cloudflare/kumo/components/collapsible";
import { Input } from "@cloudflare/kumo/components/input";
import { Select } from "@cloudflare/kumo/components/select";
import { SensitiveInput } from "@cloudflare/kumo/components/sensitive-input";
import { useKumoToastManager } from "@cloudflare/kumo/components/toast";
import { WarningCircleIcon } from "@phosphor-icons/react";
import { ApiError } from "@/api/client";
import {
  configSchema,
  entriesToServices,
  servicesToEntries,
  updateConfig,
  type Config,
} from "@/api/config";
import { FormSubmitButton } from "@/components/FormSubmitButton";
import { ConfigServicesEditor } from "./ConfigServicesEditor";

type ConfigFormState = {
  errorMessage: string | null;
  fieldErrors: Record<string, string>;
  result: Config | null;
  completedAt: number;
};

const initialActionState: ConfigFormState = {
  errorMessage: null,
  fieldErrors: {},
  result: null,
  completedAt: 0,
};

function fieldError(message: string | undefined): { error: string } | object {
  return message ? { error: message } : {};
}

function parseConfigPayload(formData: FormData): Config {
  const payload = formData.get("payload");
  if (typeof payload !== "string") {
    throw new TypeError("Config payload is missing.");
  }

  return configSchema.parse(JSON.parse(payload) as unknown);
}

export function ConfigForm(props: Readonly<{
  config: Config;
  protocols: string[];
  onSaved?: (config: Config) => void;
}>) {
  const toast = useKumoToastManager();
  const handledAtRef = useRef(0);
  const [draft, setDraft] = useState<Config>(props.config);
  const [openSections, setOpenSections] = useState({
    server: true,
    shared: true,
  });
  const [state, submitAction] = useActionState(
    async (
      _previousState: ConfigFormState,
      formData: FormData,
    ): Promise<ConfigFormState> => {
      try {
        const nextConfig = parseConfigPayload(formData);
        const result = await updateConfig(nextConfig);
        return {
          errorMessage: null,
          fieldErrors: {},
          result,
          completedAt: Date.now(),
        };
      } catch (error) {
        if (error instanceof ApiError) {
          return {
            errorMessage: error.message,
            fieldErrors: error.fieldErrors ?? {},
            result: null,
            completedAt: Date.now(),
          };
        }

        return {
          errorMessage: error instanceof Error ? error.message : "Save failed.",
          fieldErrors: {},
          result: null,
          completedAt: Date.now(),
        };
      }
    },
    initialActionState,
  );

  useEffect(() => {
    setDraft(props.config);
  }, [props.config]);

  useEffect(() => {
    if (
      !state.result ||
      state.completedAt === 0 ||
      handledAtRef.current === state.completedAt
    ) {
      return;
    }

    handledAtRef.current = state.completedAt;
    toast.add({
      variant: "success",
      title: "Configuration saved",
      description:
        "Runtime settings have been updated for the current session.",
    });
    props.onSaved?.(state.result);
  }, [props.onSaved, state.completedAt, state.result, toast]);

  return (
    <form action={submitAction} className="space-y-4">
      <input
        type="hidden"
        name="payload"
        value={JSON.stringify(draft)}
        readOnly
      />

      {state.errorMessage ? (
        <Banner
          variant="error"
          icon={<WarningCircleIcon weight="fill" />}
          title="Unable to save configuration"
          description={state.errorMessage}
        />
      ) : null}

      <Collapsible.Root
        open={openSections.server}
        onOpenChange={(open) => {
          setOpenSections((current) => ({ ...current, server: open }));
        }}
      >
        <Collapsible.DefaultTrigger className=" hover:bg-kumo-overlay text-kumo-fg rounded-md px-3 py-2">
          Server
        </Collapsible.DefaultTrigger>
        <Collapsible.DefaultPanel keepMounted>
          <div className="grid gap-4 md:grid-cols-2">
            <Input
              label="Flag checker URL"
              value={draft.server.url_flag_checker}
              {...fieldError(state.fieldErrors.url_flag_checker)}
              onChange={(event) => {
                setDraft((current) => ({
                  ...current,
                  server: {
                    ...current.server,
                    url_flag_checker: event.target.value,
                  },
                }));
              }}
            />
            <SensitiveInput
              label="Team token"
              value={draft.server.team_token}
              {...fieldError(state.fieldErrors.team_token)}
              onValueChange={(value) => {
                setDraft((current) => ({
                  ...current,
                  server: {
                    ...current.server,
                    team_token: value,
                  },
                }));
              }}
            />
            <Select
              label="Protocol"
              value={draft.server.protocol}
              onValueChange={(value) => {
                setDraft((current) => ({
                  ...current,
                  server: {
                    ...current.server,
                    protocol: String(value),
                  },
                }));
              }}
            >
              {props.protocols.map((protocol) => (
                <Select.Option key={protocol} value={protocol}>
                  {protocol}
                </Select.Option>
              ))}
            </Select>
            <Input
              label="Tick time"
              type="number"
              min={1}
              value={draft.server.tick_time}
              {...fieldError(state.fieldErrors.tick_time)}
              onChange={(event) => {
                setDraft((current) => ({
                  ...current,
                  server: {
                    ...current.server,
                    tick_time: Number(event.target.value),
                  },
                }));
              }}
            />
            <Input
              label="Submit checker time"
              type="number"
              min={0}
              value={draft.server.submit_flag_checker_time}
              {...fieldError(state.fieldErrors.submit_flag_checker_time)}
              onChange={(event) => {
                setDraft((current) => ({
                  ...current,
                  server: {
                    ...current.server,
                    submit_flag_checker_time: Number(event.target.value),
                  },
                }));
              }}
            />
            <Input
              label="Max flag batch size"
              type="number"
              min={1}
              value={draft.server.max_flag_batch_size}
              {...fieldError(state.fieldErrors.max_flag_batch_size)}
              onChange={(event) => {
                setDraft((current) => ({
                  ...current,
                  server: {
                    ...current.server,
                    max_flag_batch_size: Number(event.target.value),
                  },
                }));
              }}
            />
            <Input
              label="Flag TTL"
              type="number"
              min={0}
              value={draft.server.flag_ttl}
              {...fieldError(state.fieldErrors.flag_ttl)}
              onChange={(event) => {
                setDraft((current) => ({
                  ...current,
                  server: {
                    ...current.server,
                    flag_ttl: Number(event.target.value),
                  },
                }));
              }}
            />
            <Input
              label="Start time"
              value={draft.server.start_time}
              {...fieldError(state.fieldErrors.start_time)}
              onChange={(event) => {
                setDraft((current) => ({
                  ...current,
                  server: {
                    ...current.server,
                    start_time: event.target.value,
                  },
                }));
              }}
            />
            <Input
              label="End time"
              value={draft.server.end_time}
              {...fieldError(state.fieldErrors.end_time)}
              onChange={(event) => {
                setDraft((current) => ({
                  ...current,
                  server: {
                    ...current.server,
                    end_time: event.target.value,
                  },
                }));
              }}
            />
          </div>
        </Collapsible.DefaultPanel>
      </Collapsible.Root>

      <Collapsible.Root
        open={openSections.shared}
        onOpenChange={(open) => {
          setOpenSections((current) => ({ ...current, shared: open }));
        }}
      >
        <Collapsible.DefaultTrigger className=" hover:bg-kumo-overlay text-kumo-fg rounded-md px-3 py-2">
          Shared
        </Collapsible.DefaultTrigger>
        <Collapsible.DefaultPanel keepMounted>
          <div className="grid gap-4 md:grid-cols-2">
            <Input
              label="Regex Flag"
              value={draft.shared.regex_flag}
              {...fieldError(state.fieldErrors.regex_flag)}
              onChange={(event) => {
                setDraft((current) => ({
                  ...current,
                  shared: {
                    ...current.shared,
                    regex_flag: event.target.value,
                  },
                }));
              }}
            />
            <Input
              label="IP format for teams"
              value={draft.shared.format_ip_teams}
              {...fieldError(state.fieldErrors.format_ip_teams)}
              onChange={(event) => {
                setDraft((current) => ({
                  ...current,
                  shared: {
                    ...current.shared,
                    format_ip_teams: event.target.value,
                  },
                }));
              }}
            />
            <Input
              label="Flag IDs URL"
              value={draft.shared.url_flag_ids}
              {...fieldError(state.fieldErrors.url_flag_ids)}
              onChange={(event) => {
                setDraft((current) => ({
                  ...current,
                  shared: {
                    ...current.shared,
                    url_flag_ids: event.target.value,
                  },
                }));
              }}
            />
            <Input
              label="My team ID"
              type="number"
              min={0}
              value={draft.shared.my_team_id}
              {...fieldError(state.fieldErrors.my_team_id)}
              onChange={(event) => {
                setDraft((current) => ({
                  ...current,
                  shared: {
                    ...current.shared,
                    my_team_id: Number(event.target.value),
                  },
                }));
              }}
            />
            <Input
              label="NOP team"
              type="number"
              min={0}
              value={draft.shared.nop_team}
              {...fieldError(state.fieldErrors.nop_team)}
              onChange={(event) => {
                setDraft((current) => ({
                  ...current,
                  shared: {
                    ...current.shared,
                    nop_team: Number(event.target.value),
                  },
                }));
              }}
            />
            <Input
              label="Range IP teams"
              type="number"
              min={0}
              value={draft.shared.range_ip_teams}
              {...fieldError(state.fieldErrors.range_ip_teams)}
              onChange={(event) => {
                setDraft((current) => ({
                  ...current,
                  shared: {
                    ...current.shared,
                    range_ip_teams: Number(event.target.value),
                  },
                }));
              }}
            />
            <Input
              label="Flag Ids Format"
              type="text"
              value={draft.shared.flagids_format}
              {...fieldError(state.fieldErrors.flagids_format)}
              onChange={(event) => {
                setDraft((current) => ({
                  ...current,
                  shared: {
                    ...current.shared,
                    flagids_format: event.target.value,
                  },
                }));
              }}
            />

            <div className="rounded-xl border border-kumo-line bg-kumo-overlay p-4 md:col-span-2">
              <ConfigServicesEditor
                services={servicesToEntries(draft.shared.services)}
                onChange={(services) => {
                  setDraft((current) => ({
                    ...current,
                    shared: {
                      ...current.shared,
                      services: entriesToServices(services),
                    },
                  }));
                }}
              />
              {state.fieldErrors.services ? (
                <p className="mt-2 text-sm text-kumo-danger">
                  {state.fieldErrors.services}
                </p>
              ) : null}
            </div>
          </div>
        </Collapsible.DefaultPanel>
      </Collapsible.Root>

      <div className="rounded-xl border border-kumo-line bg-kumo-overlay p-4 text-sm text-kumo-fg-secondary">
        <p>Configured: {draft.configured ? "true" : "false"}</p>
      </div>

      <div className="flex justify-end">
        <FormSubmitButton pendingLabel="Saving...">
          Save Configuration
        </FormSubmitButton>
      </div>
    </form>
  );
}
