import { useQuery } from "@tanstack/react-query";
import { getTokens } from "@/lib/tokens";
import { getMe } from "../api";

// lib/providers.tsx refetches ["me"] when another tab logs in or out. It
// can't import features, so it spells the key out; keep the two the same.
export const ME_KEY = ["me"];

// The signed-in user, or null for a guest. With no tokens saved there is no
// request. A dead session never shows up here as null: apiFetch refreshes, and
// if that fails it sends the user to /login itself.
// staleTime Infinity: only login, logout and the other-tab listener change it.
export function useMe() {
  return useQuery({
    queryKey: ME_KEY,
    queryFn: async ({ signal }) => (getTokens() ? (await getMe(signal)).user : null),
    staleTime: Infinity,
  });
}
