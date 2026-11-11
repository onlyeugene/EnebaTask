import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { autocomplete } from "../api";
import { useDebounce } from "./use-debounce";

// The API returns nothing under 2 chars, so don't ask.
export const MIN_CHARS = 2;

export function useAutocomplete(text: string) {
  // 250 ms after the last key (API.md), so "thundr" is one request, not six.
  const q = useDebounce(text.trim(), 250);
  return useQuery({
    queryKey: ["autocomplete", q],
    // The signal cancels a request whose answer is no longer wanted.
    queryFn: ({ signal }) => autocomplete(q, signal),
    enabled: q.length >= MIN_CHARS,
    // Keep the last list while the next one loads, so it doesn't flicker per key.
    placeholderData: keepPreviousData,
    staleTime: 60_000,
    // No retry: the next key press asks again anyway, and Enter still works.
    retry: false,
  });
}
