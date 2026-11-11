import { keepPreviousData, useQuery } from "@tanstack/react-query";
import type { ProductDetail } from "@/types/api";
import { listProducts } from "../api";

// Up to `limit` other products with the same type and platform, best sellers
// first (the store's default sort). Asks for one extra, because the current
// product may be in the list and is dropped.
export function useAlsoViewed(product: ProductDetail, limit: number) {
  const query = useQuery({
    queryKey: ["also-viewed", product.type, product.platform, limit],
    queryFn: ({ signal }) =>
      listProducts(
        { type: product.type, platform: product.platform, pageSize: String(limit + 1) },
        signal,
      ),
    // "Load more" keeps the cards on screen while the bigger list loads.
    placeholderData: keepPreviousData,
  });

  const items = (query.data?.items ?? []).filter((p) => p.id !== product.id).slice(0, limit);
  // `total` counts the current product too when it is in stock (the list's
  // default is in-stock only), so take it off before asking "is there more?".
  const others = (query.data?.total ?? 0) - (product.minPriceCents !== null ? 1 : 0);
  return { items, hasMore: others > limit, loading: query.isFetching };
}
