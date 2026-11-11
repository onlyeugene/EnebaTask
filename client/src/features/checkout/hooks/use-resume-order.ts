import { useEffect } from "react";
import { ApiError } from "@/lib/client";
import { getTokens } from "@/lib/tokens";
import { getOrder } from "@/features/orders/api";
import { clearHeldOrder, getHeldOrder } from "../held-order";
import { useShowOrder } from "./use-show-order";

// Run by checkout and the cart page. If this tab left an order held (a
// refresh or crash mid-payment), reload it once:
// - paid: the payment went through but its answer was lost → its key page
// - still pending: keep it held, so Pay finishes it instead of reserving again
// - failed, expired (GET reports a lapsed hold as expired), or 404: clear it
// A network error keeps it, to check again on the next visit or Pay.
// Nothing is drawn while it checks; the page shows as usual.
export function useResumeOrder() {
  const showOrder = useShowOrder();

  useEffect(() => {
    const id = getHeldOrder();
    // Signed out (logged out in this tab since): GET would bounce to /login.
    if (!id || !getTokens()) return;

    let live = true;
    getOrder(id)
      .then(({ order }) => {
        if (!live || getHeldOrder() !== id) return; // left the page, or Pay moved on
        if (order.status === "paid") showOrder(order);
        else if (order.status !== "pending_payment") clearHeldOrder();
      })
      .catch((e) => {
        // 404: not this user's order, or not an order. Other errors may pass.
        if (e instanceof ApiError && e.status < 500) clearHeldOrder();
      });
    return () => {
      live = false;
    };
    // Once per visit. showOrder is a new function each render.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);
}
