import { apiFetch } from "@/lib/client";
import type { ProductList, ProductPage } from "@/types/api";

export function getProduct(slug: string, signal?: AbortSignal) {
  return apiFetch<ProductPage>(`/api/products/${encodeURIComponent(slug)}`, { signal });
}

// The store's list endpoint, for the gift card value tiles and "Gamers also
// viewed". `params` is built by the caller, e.g. { type, platform, pageSize }.
export function listProducts(params: Record<string, string>, signal?: AbortSignal) {
  const query = new URLSearchParams(params).toString();
  return apiFetch<ProductList>(`/api/products?${query}`, { signal });
}
