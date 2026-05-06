import useSWR, { mutate } from "swr";
import { z } from "zod";
import { apiFetch } from "./client";

export const configServicesSchema = z.record(
  z.string().trim().min(1),
  z.number().int().min(1).max(65535),
);

export const serverConfigSchema = z.object({
  url_flag_checker: z.string().trim(),
  team_token: z.string().trim(),
  submit_flag_checker_time: z.number().int().nonnegative(),
  max_flag_batch_size: z.number().int().positive(),
  protocol: z.string().trim(),
  tick_time: z.number().int().positive(),
  flag_ttl: z.number().int().nonnegative(),
  start_time: z.string().trim(),
  end_time: z.string().trim(),
});

export const sharedConfigSchema = z.object({
  services: configServicesSchema,
  regex_flag: z.string().trim(),
  format_ip_teams: z.string().trim(),
  my_team_id: z.number().int().nonnegative(),
  url_flag_ids: z.string().trim(),
  nop_team: z.number().int().nonnegative(),
  range_ip_teams: z.number().int().nonnegative().max(255),
  flagids_format: z.string().trim(),
  configured: z.boolean().default(false),
});

export const configSchema = z.object({
  server: serverConfigSchema,
  shared: sharedConfigSchema,
  configured: z.boolean().default(false),
});

export type Config = z.infer<typeof configSchema>;
export type ServerConfig = z.infer<typeof serverConfigSchema>;
export type SharedConfig = z.infer<typeof sharedConfigSchema>;
export type ConfigServices = z.infer<typeof configServicesSchema>;

export const protocolsResponseSchema = z.object({
  protocols: z.array(z.string().trim().min(1)),
});

export type ProtocolsResponse = z.infer<typeof protocolsResponseSchema>;

export const configKey = "/config/full";
export const protocolsKey = "/protocols";

export async function fetchConfig(): Promise<Config> {
  return apiFetch(configKey, {}, configSchema);
}

export function useConfig() {
  const { data } = useSWR(configKey, fetchConfig, { suspense: true });
  return data as Config;
}

export async function updateConfig(config: Config): Promise<Config> {
  await apiFetch("/config", {
    method: "POST",
    body: JSON.stringify({
      config: {
        ...config,
        configured: true,
        shared: {
          ...config.shared,
          configured: true,
        },
      },
    }),
  });

  const nextConfig = await fetchConfig();
  mutate(configKey, nextConfig, { revalidate: false });
  return nextConfig;
}

export async function fetchProtocols(): Promise<ProtocolsResponse> {
  return apiFetch(protocolsKey, {}, protocolsResponseSchema);
}

export function useProtocols() {
  const { data } = useSWR(protocolsKey, fetchProtocols, { suspense: true });
  return data as ProtocolsResponse;
}

export function servicesToEntries(
  services: ConfigServices,
): Array<[string, number]> {
  return Object.entries(services).sort(([left], [right]) =>
    left.localeCompare(right),
  );
}

export function entriesToServices(
  entries: Array<[string, number]>,
): ConfigServices {
  return Object.fromEntries(
    entries
      .map(([name, port]) => [name.trim(), port] as const)
      .filter(([name]) => name.length > 0),
  );
}
