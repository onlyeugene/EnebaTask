import { useQuery } from "@tanstack/react-query";
import { getFilters } from "../api";

// `enabled` lets the categories drawer wait until it is first opened.
export function useFilters(enabled = true) {
  // The filter lists change only when stock does, so 5 minutes is fresh enough.
  return useQuery({ queryKey: ["filters"], queryFn: getFilters, staleTime: 5 * 60_000, enabled });
}
