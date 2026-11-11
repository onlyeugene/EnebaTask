import { apiFetch } from "@/lib/client";
import type { Suggestion } from "@/types/api";

export async function autocomplete(q: string, signal?: AbortSignal) {
  const res = await apiFetch<{ suggestions: Suggestion[] }>(
    `/api/search/autocomplete?q=${encodeURIComponent(q)}`,
    { signal },
  );
  return res.suggestions;
}
