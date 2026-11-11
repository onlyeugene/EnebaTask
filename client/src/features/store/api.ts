import { apiFetch } from "@/lib/client";
import type { Filters, ProductList } from "@/types/api";

// `query` is the store page's own URL query string, passed through as is.
// The API checks every value and answers a bad one with a 400.
export function getProducts(query: string, signal?: AbortSignal) {
  return apiFetch<ProductList>(`/api/products${query ? `?${query}` : ""}`, { signal });
}

export function getFilters() {
  return apiFetch<Filters>("/api/products/filters");
}
