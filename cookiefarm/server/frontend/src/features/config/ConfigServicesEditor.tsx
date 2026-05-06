import { Button } from "@cloudflare/kumo/components/button";
import { Input } from "@cloudflare/kumo/components/input";
import { MinusIcon, PlusIcon } from "@phosphor-icons/react";

export function ConfigServicesEditor(props: Readonly<{
  services: Array<[string, number]>;
  onChange: (services: Array<[string, number]>) => void;
}>) {
  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between gap-3">
        <div>
          <p className="text-sm font-medium text-kumo-fg-primary">Services</p>
          <p className="text-xs text-kumo-fg-secondary">
            Shared service-to-port mapping from `config.yml`.
          </p>
        </div>
        <Button
          type="button"
          size="sm"
          variant="secondary"
          onClick={() => {
            props.onChange([
              ...props.services,
              ["", 8080],
            ]);
          }}
        >
          <PlusIcon size={16} />
          Add service
        </Button>
      </div>

      <div className="space-y-3">
        {props.services.map(([name, port], index) => (
          <div
            key={`${name}-${index}`}
            className="grid gap-3 rounded-xl border border-kumo-line bg-kumo-overlay p-3 md:grid-cols-[minmax(0,1fr)_180px_auto]"
          >
            <Input
              label="Name"
              value={name}
              onChange={(event) => {
                props.onChange(
                  props.services.map((currentService, currentIndex) =>
                    currentIndex === index
                      ? [event.target.value, currentService[1]]
                      : currentService,
                  ),
                );
              }}
            />

            <Input
              label="Port"
              type="number"
              min={1}
              max={65535}
              value={port}
              onChange={(event) => {
                props.onChange(
                  props.services.map((currentService, currentIndex) =>
                    currentIndex === index
                      ? [currentService[0], Number(event.target.value)]
                      : currentService,
                  ),
                );
              }}
            />

            <div className="flex items-end justify-end">
              <Button
                type="button"
                size="sm"
                variant="secondary"
                disabled={props.services.length === 1}
                onClick={() => {
                  props.onChange(
                    props.services.filter((_, currentIndex) => currentIndex !== index),
                  );
                }}
              >
                <MinusIcon size={16} />
              </Button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
