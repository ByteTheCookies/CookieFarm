import type { ReactNode } from "react";

export function ChartCard(props: Readonly<{
  title: string;
  description?: string;
  children: ReactNode;
  className?: string;
}>) {
  return (
    <section
      className={`rounded-2xl border border-kumo-line bg-kumo-base p-4 ${props.className ?? ""}`}
    >
      <h2 className={props.description ? "mb-2 text-sm font-medium text-kumo-fg-primary" : "mb-4 text-sm font-medium text-kumo-fg-primary"}>
        {props.title}
      </h2>
      {props.description ? (
        <p className="mb-4 text-sm text-kumo-fg-secondary">{props.description}</p>
      ) : null}
      {props.children}
    </section>
  );
}
