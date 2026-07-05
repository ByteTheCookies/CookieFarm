import { useEffect, useRef, useState } from "react";
import { Button } from "@cloudflare/kumo/components/button";
import { Input } from "@cloudflare/kumo/components/input";
import { MinusIcon, PlusIcon } from "@phosphor-icons/react";

export function ConfigServicesEditor(props: Readonly<{
  services: Array<[string, number]>;
  onChange: (services: Array<[string, number]>) => void;
}>) {
  const [localEntries, setLocalEntries] = useState<Array<[string, number]>>(
    () => props.services,
  );

  const ownLengthRef = useRef(props.services.length);

  useEffect(() => {
    if (props.services.length !== ownLengthRef.current) {
      ownLengthRef.current = props.services.length;
      setLocalEntries(props.services);
    }
  }, [props.services]);

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
            const newEntries: Array<[string, number]> = [
              ...localEntries,
              ["service_name", 8080],
            ];
            ownLengthRef.current = newEntries.length;
            setLocalEntries(newEntries);
            props.onChange(newEntries);
          }}
        >
          <PlusIcon size={16} />
          Add service
        </Button>
      </div>

      <div className="space-y-3">
        {localEntries.map(([name, port], index) => (
          <div
            key={index}
            className="grid gap-3 rounded-xl border border-kumo-line bg-kumo-overlay p-3 md:grid-cols-[minmax(0,1fr)_180px_auto]"
          >
            <Input
              label="Name"
              value={name}
              {...(name === "" ? { error: "Name is required" } : {})}
              onChange={(event) => {
                const newName = event.target.value;
                const newEntries = localEntries.map((entry, i) =>
                  i === index ? [newName, entry[1]] as [string, number] : entry,
                );
                setLocalEntries(newEntries);
                if (newName.trim() !== "") {
                  props.onChange(newEntries);
                }
              }}
            />

            <Input
              label="Port"
              type="number"
              min={1}
              max={65535}
              value={port}
              onChange={(event) => {
                const newPort = Number(event.target.value);
                const newEntries = localEntries.map((entry, i) =>
                  i === index ? [entry[0], newPort] as [string, number] : entry,
                );
                setLocalEntries(newEntries);
                props.onChange(newEntries);
              }}
            />

            <div className="flex items-end justify-end">
              <Button
                type="button"
                size="sm"
                variant="secondary"
                disabled={localEntries.length === 1}
                onClick={() => {
                  const newEntries = localEntries.filter((_, i) => i !== index);
                  ownLengthRef.current = newEntries.length;
                  setLocalEntries(newEntries);
                  props.onChange(newEntries);
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
