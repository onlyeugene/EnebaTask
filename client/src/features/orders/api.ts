import { apiFetch } from "@/lib/client";
import type { Order } from "@/types/api";

// Order history, newest first. No key codes in the list (API.md).
export function getOrders(signal?: AbortSignal) {
  return apiFetch<{ items: Order[] }>("/api/orders", { signal });
}

// One order, with its key codes once paid.
export function getOrder(id: string, signal?: AbortSignal) {
  return apiFetch<{ order: Order }>(`/api/orders/${id}`, { signal });
}
