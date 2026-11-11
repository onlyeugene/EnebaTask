import type { ErrorBody, TokenPair } from "@/types/api";
import { API_URL } from "./api-url";
import { clearTokens, getTokens, setTokens } from "./tokens";

// Thrown for every non-2xx response. `body` keeps the whole JSON, because some
// errors carry more than `error`: 402 has `order`, 409 OUT_OF_STOCK has `details`.
export class ApiError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string,
    public fields?: Record<string, string>,
    public body?: unknown,
  ) {
    super(message);
  }
}

type Options = { method?: string; body?: unknown; signal?: AbortSignal };

// These calls never retry after a refresh: a 401 from them means bad
// credentials or a dead refresh token, not an expired access token.
// /me and /logout are not listed: they use the access token like any other call.
const NO_REFRESH = ["/api/auth/login", "/api/auth/register", "/api/auth/refresh"];

export async function apiFetch<T>(path: string, opts: Options = {}): Promise<T> {
  const sentWith = getTokens()?.accessToken ?? null;
  let res = await send(path, opts, sentWith);

  if (res.status === 401 && sentWith && !NO_REFRESH.includes(path)) {
    if (await refreshOnce(sentWith)) {
      res = await send(path, opts, getTokens()?.accessToken ?? null); // retry once
    }
  }
  return parse<T>(res);
}

function send(path: string, opts: Options, accessToken: string | null) {
  const headers: Record<string, string> = {};
  if (opts.body !== undefined) headers["Content-Type"] = "application/json";
  if (accessToken) headers["Authorization"] = `Bearer ${accessToken}`;
  return fetch(API_URL + path, {
    method: opts.method ?? "GET",
    headers,
    body: opts.body === undefined ? undefined : JSON.stringify(opts.body),
    signal: opts.signal,
  });
}

async function parse<T>(res: Response): Promise<T> {
  if (res.status === 204) return undefined as T;
  const data = await res.json().catch(() => null);
  if (res.ok) return data as T;
  const err = (data as ErrorBody | null)?.error;
  throw new ApiError(
    res.status,
    err?.code ?? "UNKNOWN",
    err?.message ?? `Request failed (${res.status})`,
    err?.fields,
    data,
  );
}

// Single flight: while one refresh is running, every other 401 in this tab
// waits on the same promise. 5 parallel 401s cause 1 refresh, not 5 (a second
// use of the same refresh token would trip the server's reuse alarm and log
// the user out everywhere).
let pending: Promise<boolean> | null = null;

function refreshOnce(staleAccess: string): Promise<boolean> {
  if (!pending) {
    pending = refresh(staleAccess).finally(() => {
      pending = null;
    });
  }
  return pending;
}

async function refresh(staleAccess: string): Promise<boolean> {
  // The Web Lock makes tabs take turns. Inside it we read the tokens again:
  // if they changed, another tab (or an earlier refresh here) already rotated
  // them, so we use the new ones instead of sending the old refresh token.
  const run = async (): Promise<boolean> => {
    const current = getTokens();
    if (!current) return false;
    if (current.accessToken !== staleAccess) return true;

    const res = await fetch(`${API_URL}/api/auth/refresh`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refreshToken: current.refreshToken }),
    });
    // Only a 401 means the session is over. A 429 or 5xx is thrown, so a
    // server blip shows an error instead of logging the user out.
    if (res.status === 401) return false;
    const pair = await parse<TokenPair>(res);
    setTokens(pair);
    return true;
  };

  const ok = navigator.locks ? await navigator.locks.request("auth-refresh", run) : await run();
  if (!ok) sessionLost();
  return ok;
}

// A full page load also drops the in-memory query cache, so no user data
// from the old session stays on screen.
function sessionLost() {
  clearTokens();
  const next = encodeURIComponent(window.location.pathname + window.location.search);
  window.location.href = `/login?next=${next}`;
}
