import useSWR, { mutate, type SWRConfiguration } from "swr";
import { z } from "zod";
import { apiFetch } from "./client";

export const flagStatusSchema = z.union([
  z.literal(0),
  z.literal(1),
  z.literal(2),
  z.literal(3),
  z.literal(4),
]);

export type FlagStatus = z.infer<typeof flagStatusSchema>;

export const flagSchema = z.object({
  flag_code: z.string(),
  service_name: z.string(),
  port_service: z.number().int().nonnegative(),
  submit_time: z.number().int().nonnegative(),
  response_time: z.number().int().nonnegative(),
  msg: z.string(),
  status: flagStatusSchema,
  team_id: z.number().int(),
  username: z.string(),
  exploit_name: z.string(),
});

export type Flag = z.infer<typeof flagSchema>;

export const flagsResponseSchema = z.object({
  flags: z.array(flagSchema),
  n_flags: z.number().int().nonnegative(),
  next: z.string(),
});

export type FlagsResponse = z.infer<typeof flagsResponseSchema>;

export type FlagsQuery = {
  limit: number;
  cursor?: string;
  status?: FlagStatus;
  service?: string;
  team?: string;
  search?: string;
  searchField?: string;
};

const submitFlagRequestSchema = z.object({
  flag: flagSchema,
});

function buildFlagsQuery(params: FlagsQuery): string {
  const query = new URLSearchParams();
  if (params.cursor) {
    query.set("cursor", params.cursor);
  }

  if (params.status !== undefined) {
    query.set("status", String(params.status));
  }
  if (params.service) {
    query.set("service", params.service);
  }
  if (params.team) {
    query.set("team", params.team);
  }
  if (params.search) {
    query.set("search", params.search);
  }
  if (params.searchField) {
    query.set("search_field", params.searchField);
  }

  return query.toString();
}

export function flagsKey(params: FlagsQuery) {
  return ["/flags", params.limit, buildFlagsQuery(params)] as const;
}

export const allFlagsKey = "/flags";

export async function fetchFlags(params: FlagsQuery): Promise<FlagsResponse> {
  return apiFetch(
    `/flags/${params.limit}?${buildFlagsQuery(params)}`,
    {},
    flagsResponseSchema,
  );
}

export async function fetchAllFlags(): Promise<FlagsResponse> {
  return apiFetch(allFlagsKey, {}, flagsResponseSchema);
}

export function useFlags(params: FlagsQuery, options: SWRConfiguration = {}) {
  return useSWR(flagsKey(params), () => fetchFlags(params), {
    suspense: true,
    keepPreviousData: true,
    ...options,
  });
}

export function useAllFlags(options: SWRConfiguration = {}) {
  return useSWR(allFlagsKey, fetchAllFlags, {
    suspense: true,
    ...options,
  });
}

export async function submitFlag(flag: Flag): Promise<void> {
  await apiFetch(
    "/submit-flag",
    {
      method: "POST",
      body: JSON.stringify(
        submitFlagRequestSchema.parse({
          flag,
        }),
      ),
    },
  );
  invalidateFlagsCache();
}

export async function deleteFlag(flagCode: string): Promise<void> {
  await apiFetch(
    `/delete-flag?flag=${encodeURIComponent(flagCode)}`,
    {
      method: "DELETE",
    },
  );
invalidateFlagsCache();
}

export function invalidateFlagsCache() {
  return mutate((key) => Array.isArray(key) ? key[0] === "/flags" : key === allFlagsKey);
}
