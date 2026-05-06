import { useEffect, useState } from "react";
import { Banner } from "@cloudflare/kumo/components/banner";
import { Button } from "@cloudflare/kumo/components/button";
import {
  Dialog,
  DialogClose,
  DialogRoot,
  DialogTitle,
} from "@cloudflare/kumo/components/dialog";
import { Input } from "@cloudflare/kumo/components/input";
import { cn } from "@cloudflare/kumo/utils";
import {
  Check,
  Copy,
  WarningCircleIcon,
  X,
} from "@phosphor-icons/react";

export const KUMO_DELETE_RESOURCE_VARIANTS = {
  size: {
    sm: {
      classes: "",
      description: "Small dialog for simple delete confirmations",
    },
    base: {
      classes: "",
      description: "Default delete confirmation dialog size",
    },
  },
} as const;

export const KUMO_DELETE_RESOURCE_DEFAULT_VARIANTS = {
  size: "base",
} as const;

export type KumoDeleteResourceSize =
  keyof typeof KUMO_DELETE_RESOURCE_VARIANTS.size;

export interface KumoDeleteResourceVariantsProps {
  size?: KumoDeleteResourceSize;
}

export interface DeleteResourceProps extends KumoDeleteResourceVariantsProps {
  /** Whether the dialog is open */
  open: boolean;
  /** Callback when open state changes */
  onOpenChange: (open: boolean) => void;
  /** The type of resource being deleted (e.g., "Zone", "Worker", "KV Namespace") */
  resourceType: string;
  /** The name of the specific resource being deleted */
  resourceName: string;
  /** Callback when delete is confirmed */
  onDelete: () => void | Promise<void>;
  /** Whether the delete action is in progress */
  isDeleting?: boolean;
  /** Whether the confirmation input should be case-sensitive (default: true) */
  caseSensitive?: boolean;
  /** Custom delete button text (defaults to "Delete {resourceType}") */
  deleteButtonText?: string;
  /** Additional className for the dialog */
  className?: string;
  /** Error message to display if the delete action fails */
  errorMessage?: string;
}

export function DeleteResource({
  open,
  onOpenChange,
  resourceType,
  resourceName,
  onDelete,
  isDeleting = false,
  caseSensitive = true,
  deleteButtonText,
  size = KUMO_DELETE_RESOURCE_DEFAULT_VARIANTS.size,
  errorMessage,
  className,
}: Readonly<DeleteResourceProps>) {
  const [confirmationInput, setConfirmationInput] = useState("");
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (!open) {
      setConfirmationInput("");
      setCopied(false);
    }
  }, [open]);

  const normalizeForComparison = (str: string) =>
    caseSensitive ? str : str.toLowerCase();

  const isConfirmed =
    normalizeForComparison(confirmationInput) ===
    normalizeForComparison(resourceName);

  async function handleDelete() {
    if (!isConfirmed || isDeleting) return;
    await onDelete();
  }

  async function handleCopy() {
    await navigator.clipboard.writeText(resourceName);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  }

  return (
    <DialogRoot open={open} onOpenChange={onOpenChange}>
      <Dialog size={size} className={cn("p-0", className)}>
        <div className="flex items-center justify-between border-b border-kumo-line px-6 py-4">
          <DialogTitle className="text-lg font-semibold">
            Delete {resourceName}
          </DialogTitle>
          <DialogClose
            render={(props) => (
              <button
                {...props}
                type="button"
                aria-label="Close"
                disabled={isDeleting}
                className="inline-flex size-6.5 items-center justify-center rounded-md text-kumo-default hover:bg-kumo-tint disabled:opacity-50"
              >
                <X size={18} />
              </button>
            )}
          />
        </div>

        <div className="flex flex-col p-6 gap-4">
          <div className="flex flex-col gap-2">
            {errorMessage && (
              <Banner icon={<WarningCircleIcon />} variant="error">
                {errorMessage}
              </Banner>
            )}
            <p className="text-base text-kumo-subtle max-w-prose text-pretty">
              This action cannot be undone. This will permanently delete the{" "}
              <span className="font-medium text-kumo-default">
                {resourceName}
              </span>{" "}
              {resourceType.toLowerCase()}.
            </p>
          </div>

          <div className="flex flex-col gap-2">
            <div className="flex items-center gap-1.5 text-base">
              <span>
                Type{" "}
                <button
                  className="font-mono text-sm inline font-semibold bg-kumo-tint hover:bg-kumo-fill rounded-md px-2 py-1 group hover:cursor-pointer"
                  onClick={handleCopy}
                  aria-label={`Copy ${resourceName} to clipboard`}
                >
                  {resourceName}

                  {copied ? (
                    <Check
                      size={12}
                      weight="bold"
                      className="inline ml-1.5"
                    />
                  ) : (
                    <Copy
                      size={12}
                      weight="bold"
                      className="inline text-kumo-subtle group-hover:text-kumo-default ml-1.5"
                    />
                  )}
                </button>{" "}
                to confirm:
              </span>
            </div>
            <Input
              placeholder={resourceName}
              value={confirmationInput}
              onChange={(e) => setConfirmationInput(e.target.value)}
              disabled={isDeleting}
              autoComplete="off"
              autoCorrect="off"
              autoCapitalize="off"
              spellCheck={false}
              aria-label={`Type ${resourceName} to confirm deletion`}
              className="w-full"
            />
          </div>
        </div>

        <div className="flex justify-end gap-3 border-t border-kumo-line px-6 py-4">
          <DialogClose
            render={(props) => (
              <button
                {...props}
                type="button"
                disabled={isDeleting}
                className="inline-flex h-9 items-center rounded-lg bg-kumo-base px-3 text-base text-kumo-default ring ring-kumo-hairline hover:bg-kumo-tint disabled:opacity-50"
              >
                Cancel
              </button>
            )}
          />
          <Button
            variant="destructive"
            onClick={handleDelete}
            disabled={!isConfirmed || isDeleting}
            loading={isDeleting}
          >
            {deleteButtonText || `Delete ${resourceType}`}
          </Button>
        </div>
      </Dialog>
    </DialogRoot>
  );
}

DeleteResource.displayName = "DeleteResource";
