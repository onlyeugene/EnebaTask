import { apiFetch } from "@/lib/client";
import type { Wishlist, WishlistMergeInput } from "@/types/api";

// The signed-in user's wishlist in the query cache. lib/providers.tsx spells
// it out too (it can't import features); keep the two the same.
export const WISHLIST_KEY = ["wishlist"];

export function getWishlist(signal?: AbortSignal) {
  return apiFetch<Wishlist>("/api/wishlist", { signal });
}

export function addToWishlist(productId: string) {
  return apiFetch<Wishlist>(`/api/wishlist/${productId}`, { method: "PUT" });
}

export function removeFromWishlist(productId: string) {
  return apiFetch<Wishlist>(`/api/wishlist/${productId}`, { method: "DELETE" });
}

export function mergeWishlist(items: WishlistMergeInput[]) {
  return apiFetch<Wishlist>("/api/wishlist/merge", { method: "POST", body: { items } });
}
