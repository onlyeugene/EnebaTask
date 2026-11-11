import { apiFetch } from "@/lib/client";
import type { Order, PayInput } from "@/types/api";

// Reserves keys for everything in the signed-in user's server cart.
// The body is empty: the server reads the cart itself (API.md, Orders).
export function createOrder() {
  return apiFetch<{ order: Order }>("/api/orders", { method: "POST", body: {} });
}

// 200 → paid, with keys. 402 → declined: the ApiError's body carries the failed order.
export function payOrder(id: string, card: PayInput) {
  return apiFetch<{ order: Order }>(`/api/orders/${id}/pay`, { method: "POST", body: card });
}
