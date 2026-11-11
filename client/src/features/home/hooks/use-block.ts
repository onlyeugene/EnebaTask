import { useQuery } from "@tanstack/react-query";
import { getBlock } from "../api";

export function useBlock<D>(id: string, fail: boolean) {
  return useQuery({
    // `fail` is in the key so the forced-fail request never shares a cache
    // entry with the normal one.
    queryKey: ["homepage-block", id, fail],
    queryFn: () => getBlock<D>(id, fail),
    // One retry, not the default two: a broken block should show its error
    // and Retry button within a second or two, not sit on its skeleton.
    retry: 1,
  });
}
