import { apiFetch } from "@/lib/client";
import type { Cart, CartItemInput, MergeResult } from "@/types/api";

// The signed-in user's cart in the query cache. Starts with "cart" so
// invalidating ["cart"] (lib/providers.tsx, on login in another tab) covers it.
export const SERVER_CART_KEY = ["cart", "server"];

// Live prices and stock for the guest cart. Nothing is saved on the server.
export function quoteCart(items: CartItemInput[], signal?: AbortSignal) {
  return apiFetch<Cart>("/api/cart/quote", { method: "POST", body: { items }, signal });
}

export function getCart(signal?: AbortSignal) {
  return apiFetch<Cart>("/api/cart", { signal });
}

// Sets the quantity (does not add to it).
export function upsertItem(offerId: string, quantity: number) {
  return apiFetch<Cart>(`/api/cart/items/${offerId}`, { method: "PUT", body: { quantity } });
}

export function removeItem(offerId: string) {
  return apiFetch<Cart>(`/api/cart/items/${offerId}`, { method: "DELETE" });
}

export function mergeCart(items: CartItemInput[]) {
  return apiFetch<MergeResult>("/api/cart/merge", { method: "POST", body: { items } });
}
