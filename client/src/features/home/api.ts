import { apiFetch } from "@/lib/client";
import type { Block, LayoutEntry } from "@/types/api";

export function getHomepageLayout() {
  return apiFetch<{ blocks: LayoutEntry[] }>("/api/homepage");
}

// `fail` adds ?fail=1, which the API (development only) answers with a 503,
// to show one broken block among working ones.
export function getBlock<D>(id: string, fail: boolean) {
  const query = fail ? "?fail=1" : "";
  return apiFetch<Block<D>>(`/api/homepage/blocks/${encodeURIComponent(id)}${query}`);
}
