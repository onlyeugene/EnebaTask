import { useRouter } from "next/navigation";
import { useQueryClient } from "@tanstack/react-query";
import type { Order } from "@/types/api";
import { orderKey } from "@/features/orders/hooks/use-order";
import { clearHeldOrder } from "../held-order";

// Opens a paid order's key page. Used after Pay, and by useResumeOrder when a
// reload finds the held order already paid.
//
// The order goes in the cache first, so the keys show with no extra request.
// Cart and history are marked stale but not refetched: a refetch now would
// flash "Your cart is empty" before the redirect. The next page that shows
// them loads them fresh. replace, not push: Back shouldn't return to a used checkout.
export function useShowOrder() {
  const router = useRouter();
  const client = useQueryClient();

  return (order: Order) => {
    clearHeldOrder();
    client.setQueryData(orderKey(order.id), order);
    markChanged(client);
    router.replace(`/orders/${order.id}`);
  };
}

export function markChanged(client: ReturnType<typeof useQueryClient>) {
  void client.invalidateQueries({ queryKey: ["cart"], refetchType: "none" });
  void client.invalidateQueries({ queryKey: ["orders"], refetchType: "none" });
}
