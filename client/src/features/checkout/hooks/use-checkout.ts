import { useRouter } from "next/navigation";
import { useQueryClient } from "@tanstack/react-query";
import type { UseFormSetError } from "react-hook-form";
import type { Order } from "@/types/api";
import { ApiError } from "@/lib/client";
import { getOrder } from "@/features/orders/api";
import { createOrder, payOrder } from "../api";
import { clearHeldOrder, getHeldOrder, setHeldOrder } from "../held-order";
import { cardDigits, type PaymentValues } from "../schemas";
import { markChanged, useShowOrder } from "./use-show-order";

// The "Pay" button: reserve keys (POST /api/orders), then pay (POST /pay).
//
// The held order (held-order.ts) is the one whose payment we may still need
// to finish. It is kept when the outcome is unknown (network error, 5xx) or
// the card was rejected for its format (422), so the next click pays the same
// order instead of reserving a second set of keys. It is cleared once the
// order is final (paid, declined, expired), so the next click starts fresh.
// It lives in sessionStorage, so a refresh mid-payment keeps it too
// (useResumeOrder checks it when the page opens again).
//
// Lost response (PLAN.md phase 15): the pay may commit on the server while its
// answer never arrives. The next click pays the same order again, gets
// 409 ORDER_NOT_PAYABLE, reloads the order, finds it paid and shows the keys.
export function useCheckout() {
  const router = useRouter();
  const client = useQueryClient();
  const showOrder = useShowOrder();

  return async (v: PaymentValues, setError: UseFormSetError<PaymentValues>) => {
    let id = getHeldOrder();
    if (!id) {
      try {
        id = (await createOrder()).order.id;
        setHeldOrder(id);
      } catch (e) {
        createFailed(e, setError);
        // 409 OUT_OF_STOCK or 422 empty cart: reload the cart, so the line at
        // fault shows its problem and the Pay button is blocked.
        void client.invalidateQueries({ queryKey: ["cart"] });
        return;
      }
    }

    try {
      showOrder((await payOrder(id, { ...v, cardNumber: cardDigits(v.cardNumber) })).order);
    } catch (e) {
      if (!(e instanceof ApiError)) {
        setError("root", { message: "Couldn't reach the server. Press Pay again: you won't be charged twice." });
        return;
      }
      if (e.code === "PAYMENT_DECLINED") {
        // The server released the keys and left the cart as it was. Back to
        // the cart (step 1), which says what happened.
        clearHeldOrder();
        markChanged(client);
        router.replace("/cart?payment=declined");
        return;
      }
      if (e.code === "ORDER_NOT_PAYABLE") {
        await notPayable(id, setError);
        return;
      }
      if (e.fields) {
        // The server's messages are short ("must be 3 or 4 digits"), so the
        // field's name goes in front. The order stays held: fix the card, pay it.
        for (const [name, msg] of Object.entries(e.fields)) {
          setError(name as keyof PaymentValues, { message: `${FIELD_LABELS[name] ?? name} ${msg}` });
        }
        return;
      }
      // A 5xx may have committed, so keep the order and let the next click
      // find out (the 409 path above). Any other 4xx (404) won't change on retry.
      if (e.status < 500) clearHeldOrder();
      setError("root", { message: "Something went wrong. Press Pay to try again." });
    }
  };

  // 409 on pay: either an earlier pay of this order went through (show the
  // keys) or the 10-minute hold ran out (start over on the next click).
  async function notPayable(id: string, setError: UseFormSetError<PaymentValues>) {
    let order: Order;
    try {
      order = (await getOrder(id)).order;
    } catch {
      setError("root", { message: "Couldn't check your order. Press Pay to try again." });
      return;
    }
    if (order.status === "paid") {
      showOrder(order);
      return;
    }
    clearHeldOrder();
    setError("root", { message: "Your reservation ran out. Press Pay to reserve your keys again." });
  }
}

const FIELD_LABELS: Record<string, string> = {
  cardNumber: "Card number",
  expiry: "Expiry",
  cvc: "CVC",
  name: "Name",
};

function createFailed(e: unknown, setError: UseFormSetError<PaymentValues>) {
  if (!(e instanceof ApiError)) {
    setError("root", { message: "Couldn't reach the server. Check your connection and try again." });
  } else if (e.code === "OUT_OF_STOCK") {
    setError("root", { message: "Not enough keys left for an item in your cart. Check your cart and try again." });
  } else if (e.status === 422) {
    setError("root", { message: "Your cart is empty." });
  } else {
    setError("root", { message: "Something went wrong. Please try again." });
  }
}
