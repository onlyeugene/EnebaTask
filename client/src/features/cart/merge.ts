import type { QueryClient } from "@tanstack/react-query";
import { clearHeldOrder } from "@/features/checkout/held-order";
import { mergeCart, SERVER_CART_KEY } from "./api";
import { useMergeNotice } from "./merge-notice-store";
import { useGuestCart } from "./store";

// Moves the guest cart into the signed-in user's server cart (PLAN.md 5b).
// The server applies the rules (same offer → larger quantity, capped at stock
// and 10). Called right after login or register, and again by
// useGuestCartSync when a signed-in page finds guest lines left over.
//
// The guest cart is only cleared after the server answers 200. If the request
// fails it throws and the guest lines stay, so nothing is lost and the next
// try sends them again. Sending twice is safe: "larger quantity" gives the
// same cart however many times it runs.
export async function mergeGuestCart(client: QueryClient) {
  // The login page has no header popover, so the saved cart may not be loaded yet.
  await useGuestCart.persist.rehydrate();
  const sent = useGuestCart.getState().items;
  if (sent.length === 0) return;

  clearHeldOrder(); // the server cart is about to change; see checkout/held-order.ts
  const { cart, dropped } = await mergeCart(sent);
  client.setQueryData(SERVER_CART_KEY, cart);

  // Remove only the lines that were sent. A line added or raised in another
  // tab while the request ran stays, and goes in the next merge.
  useGuestCart.setState((s) => ({
    items: s.items.filter((i) => !sent.some((x) => x.offerId === i.offerId && x.quantity >= i.quantity)),
  }));
  if (dropped.length > 0) useMergeNotice.getState().show(dropped.length);
}
