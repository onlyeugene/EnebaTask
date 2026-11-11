import { useQuery } from "@tanstack/react-query";
import { getOrders } from "../api";

export const ORDERS_KEY = ["orders"];

// The signed-in user's order history.
export function useOrders() {
  return useQuery({
    queryKey: ORDERS_KEY,
    queryFn: async ({ signal }) => (await getOrders(signal)).items,
  });
}
