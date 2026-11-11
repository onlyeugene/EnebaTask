import { queryOptions, useQuery } from "@tanstack/react-query";
import { getProduct } from "../api";

// Shared by the product page and the cards' "Add to cart" (use-add-cheapest),
// so both read the same cache entry.
export function productQuery(slug: string) {
  return queryOptions({
    queryKey: ["product", slug],
    queryFn: ({ signal }) => getProduct(slug, signal),
  });
}

export function useProduct(slug: string) {
  return useQuery(productQuery(slug));
}
