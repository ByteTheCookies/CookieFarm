import { Button } from "@cloudflare/kumo/components/button";
import { Input } from "@cloudflare/kumo/components/input";
import { MinusIcon, PlusIcon } from "@phosphor-icons/react";

export function ConfigListEditor(props: Readonly<{
  label: string;
  items: string[];
  placeholder: string;
  onChange: (items: string[]) => void;
}>) {
  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between gap-3">
        <div>
          <p className="text-sm font-medium text-kumo-fg-primary">{props.label}</p>
          <p className="text-xs text-kumo-fg-secondary">One entry per operator-controlled row.</p>
        </div>
        <Button
          type="button"
          size="sm"
          variant="secondary"
          onClick={() => {
            props.onChange([...props.items, ""]);
          }}
        >
          <PlusIcon size={16} />
          Add row
        </Button>
      </div>

      <div className="space-y-2">
        {props.items.map((item, index) => (
          <div key={`${props.label}-${index}`} className="flex items-center gap-2">
            <Input
              value={item}
              placeholder={props.placeholder}
              onChange={(event) => {
                props.onChange(
                  props.items.map((currentItem, currentIndex) =>
                    currentIndex === index ? event.target.value : currentItem,
                  ),
                );
              }}
            />
            <Button
              type="button"
              size="sm"
              variant="secondary"
              disabled={props.items.length === 1}
              onClick={() => {
                props.onChange(props.items.filter((_, currentIndex) => currentIndex !== index));
              }}
            >
              <MinusIcon size={16} />
            </Button>
          </div>
        ))}
      </div>
    </div>
  );
}
