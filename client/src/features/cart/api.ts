import { apiFetch } from "@/lib/client";
import type { Cart, CartItemInput } from "@/types/api";

// Live prices and stock for the guest cart. Nothing is saved on the server.
// The signed-in calls (getCart, upsertItem, removeItem, mergeCart) come with
// login in phase 14.
export function quoteCart(items: CartItemInput[], signal?: AbortSignal) {
  return apiFetch<Cart>("/api/cart/quote", { method: "POST", body: { items }, signal });
}
