import type { TokenPair } from "@/types/api";

// Both tokens sit under one key, so a read always gets a matching pair, even
// while another tab is writing a new one.
const KEY = "eugene.tokens";

export function getTokens(): TokenPair | null {
  if (typeof window === "undefined") return null;
  try {
    const raw = localStorage.getItem(KEY);
    return raw ? (JSON.parse(raw) as TokenPair) : null;
  } catch {
    return null;
  }
}

export function setTokens(pair: TokenPair) {
  localStorage.setItem(KEY, JSON.stringify(pair));
}

export function clearTokens() {
  localStorage.removeItem(KEY);
}
