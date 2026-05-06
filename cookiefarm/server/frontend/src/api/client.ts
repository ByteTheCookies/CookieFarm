import { z, type ZodType } from "zod";

export const API_BASE = (import.meta.env.VITE_API_BASE_URL ?? "/api/v1").trim();

export class ApiError extends Error {
  readonly status: number;
  readonly details?: string | undefined;
  readonly body?: unknown;
  readonly fieldErrors?: Record<string, string> | undefined;

  constructor(
    status: number,
    message: string,
    options: {
      body?: unknown;
      details?: string;
      fieldErrors?: Record<string, string>;
    } = {},
  ) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.body = options.body;
    this.details = options.details;
    this.fieldErrors = options.fieldErrors;
  }
}

function buildAbsoluteUrl(path: string): string {
  if (/^https?:\/\//.test(path)) {
    return path;
  }

  const normalizedBase = API_BASE.endsWith("/") ? API_BASE.slice(0, -1) : API_BASE;
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;
  return `${normalizedBase}${normalizedPath}`;
}

async function parseResponseBody(response: Response): Promise<unknown> {
  const text = await response.text();
  if (!text) {
    return undefined;
  }

  try {
    return JSON.parse(text) as unknown;
  } catch {
    return text;
  }
}

function getErrorPayload(body: unknown): Record<string, unknown> | undefined {
  return body && typeof body === "object" ? (body as Record<string, unknown>) : undefined;
}

function getFieldErrors(errorPayload: Record<string, unknown> | undefined): Record<string, string> | undefined {
  if (
    !errorPayload?.fieldErrors ||
    typeof errorPayload.fieldErrors !== "object" ||
    Array.isArray(errorPayload.fieldErrors)
  ) {
    return undefined;
  }

  return Object.fromEntries(
    Object.entries(errorPayload.fieldErrors).flatMap(([key, value]) =>
      typeof value === "string" ? [[key, value]] : [],
    ),
  );
}

function getErrorMessage(
  response: Response,
  errorPayload: Record<string, unknown> | undefined,
): string {
  if (typeof errorPayload?.message === "string") {
    return errorPayload.message;
  }

  if (typeof errorPayload?.error === "string") {
    return errorPayload.error;
  }

  return response.statusText || "Request failed";
}

function buildApiError(response: Response, body: unknown): ApiError {
  const errorPayload = getErrorPayload(body);
  const errorOptions: {
    body?: unknown;
    details?: string;
    fieldErrors?: Record<string, string>;
  } = {};

  if (body !== undefined) {
    errorOptions.body = body;
  }

  if (typeof errorPayload?.details === "string") {
    errorOptions.details = errorPayload.details;
  }

  const fieldErrors = getFieldErrors(errorPayload);
  if (fieldErrors) {
    errorOptions.fieldErrors = fieldErrors;
  }

  return new ApiError(response.status, getErrorMessage(response, errorPayload), errorOptions);
}

export async function apiFetch<T>(
  path: string,
  init: RequestInit = {},
  schema?: ZodType<T>,
): Promise<T> {
  const headers = new Headers(init.headers);
  headers.set("Accept", "application/json");
  if (init.body && !(init.body instanceof FormData) && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  const response = await fetch(buildAbsoluteUrl(path), {
    credentials: "include",
    ...init,
    headers,
  });

  const body = await parseResponseBody(response);
  if (!response.ok) {
    throw buildApiError(response, body);
  }

  if (!schema) {
    return body as T;
  }

  return schema.parse(body);
}

export function buildWebSocketUrl(path: string): string {
  const target = new URL(buildAbsoluteUrl(path), globalThis.location.origin);
  const protocol = target.protocol === "https:" ? "wss:" : "ws:";
  return `${protocol}//${target.host}${target.pathname}${target.search}`;
}

export function buildEventSourceUrl(path: string): string {
  return buildAbsoluteUrl(path);
}

export const messageResponseSchema = z.object({
  message: z.string(),
});
