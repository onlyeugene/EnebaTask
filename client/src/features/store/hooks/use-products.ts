import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { getProducts } from "../api";

export function useProducts(query: string) {
  return useQuery({
    queryKey: ["products", query],
    queryFn: ({ signal }) => getProducts(query, signal),
    // Keep the old grid on screen (dimmed) while a new filter loads, instead
    // of flashing skeletons on every click.
    placeholderData: keepPreviousData,
  });
}
