import type { ReactNode } from "react";
import { useFormStatus } from "react-dom";
import {
  Button,
  type ButtonProps,
} from "@cloudflare/kumo/components/button";

type FormSubmitButtonProps = Readonly<{
  children: ReactNode;
  pendingLabel?: ReactNode;
  variant?: ButtonProps["variant"];
  size?: ButtonProps["size"];
  className?: string;
}>;

export function FormSubmitButton(props: FormSubmitButtonProps) {
  const { pending } = useFormStatus();
  const buttonProps = {
    ...(props.variant ? { variant: props.variant } : {}),
    ...(props.size ? { size: props.size } : {}),
    ...(props.className ? { className: props.className } : {}),
  };

  return (
    <Button
      {...buttonProps}
      type="submit"
      loading={pending}
    >
      {pending ? props.pendingLabel ?? props.children : props.children}
    </Button>
  );
}
