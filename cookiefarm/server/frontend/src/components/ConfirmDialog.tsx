import { Button } from "@cloudflare/kumo/components/button";
import { Dialog } from "@cloudflare/kumo/components/dialog";

export function ConfirmDialog(props: Readonly<{
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description: string;
  confirmLabel: string;
  isPending?: boolean;
  onConfirm: () => void | Promise<void>;
}>) {
  const confirmProps = props.isPending ? { loading: true } : {};

  return (
    <Dialog.Root open={props.open} onOpenChange={props.onOpenChange} role="alertdialog">
      <Dialog size="sm" className="space-y-4 p-6">
        <div className="space-y-2">
          <Dialog.Title>{props.title}</Dialog.Title>
          <Dialog.Description>{props.description}</Dialog.Description>
        </div>

        <div className="flex items-center justify-end gap-3">
          <Button
            variant="secondary"
            onClick={() => {
              props.onOpenChange(false);
            }}
          >
            Cancel
          </Button>
          <Button
            variant="destructive"
            {...confirmProps}
            onClick={() => {
              void props.onConfirm();
            }}
          >
            {props.confirmLabel}
          </Button>
        </div>
      </Dialog>
    </Dialog.Root>
  );
}
