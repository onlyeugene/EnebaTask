import { useQuery } from "@tanstack/react-query";
import { getFilters } from "../api";

export function useFilters() {
  // The filter lists change only when stock does, so 5 minutes is fresh enough.
  return useQuery({ queryKey: ["filters"], queryFn: getFilters, staleTime: 5 * 60_000 });
}
