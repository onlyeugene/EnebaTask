import { useQuery } from "@tanstack/react-query";
import { getOrder } from "../api";

export const orderKey = (id: string) => ["order", id];

// One order. Checkout puts the paid order in the cache under this key before
// opening the order page, so the keys show without waiting for a request.
export function useOrder(id: string) {
  return useQuery({
    queryKey: orderKey(id),
    queryFn: async ({ signal }) => (await getOrder(id, signal)).order,
  });
}
