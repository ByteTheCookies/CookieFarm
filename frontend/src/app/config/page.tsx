'use client';

import { ServerConfigStepper } from '@/components/server-config-stepper';

export default function MatchOverview() {
  return (
    <div className="p-6">
      <div className="mx-auto max-w-4xl">
        <div className="mb-8">
          <h1 className="text-foreground text-3xl font-bold">
            Server Configuration
          </h1>
          <p className="text-muted-foreground">
            Configure your CTF server settings step by step
          </p>
        </div>
        <ServerConfigStepper />
      </div>
    </div>
  );
}
