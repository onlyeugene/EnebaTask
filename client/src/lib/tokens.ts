import type { TokenPair } from "@/types/api";

// Both tokens sit under one key, so a read always gets a matching pair, even
// while another tab is writing a new one. Exported for providers.tsx, which
// watches it to notice a login or logout in another tab.
export const TOKENS_KEY = "eugene.tokens";

export function getTokens(): TokenPair | null {
  if (typeof window === "undefined") return null;
  try {
    const raw = localStorage.getItem(TOKENS_KEY);
    return raw ? (JSON.parse(raw) as TokenPair) : null;
  } catch {
    return null;
  }
}

export function setTokens(pair: TokenPair) {
  // Only the pair: an AuthResponse also carries `user`, which isn't stored here.
  localStorage.setItem(TOKENS_KEY, JSON.stringify({ accessToken: pair.accessToken, refreshToken: pair.refreshToken }));
}

export function clearTokens() {
  localStorage.removeItem(TOKENS_KEY);
}
