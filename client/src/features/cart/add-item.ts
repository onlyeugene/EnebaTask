import type { QueryClient } from "@tanstack/react-query";
import { ApiError } from "@/lib/client";
import { getTokens } from "@/lib/tokens";
import { getCart, SERVER_CART_KEY, upsertItem } from "./api";
import type { AddOutcome } from "./added-modal-store";
import { MAX_QTY, useGuestCart } from "./store";

// Adds one of an offer to whichever cart is in use. Every add button (offer
// rows, product cards) goes through here.
// Signed in is read from the saved tokens, not the /me query, because a click
// needs the answer now. If the tokens are dead, apiFetch sends the user to /login.
export async function addItem(client: QueryClient, offerId: string): Promise<AddOutcome> {
  if (!getTokens()) return useGuestCart.getState().add(offerId);

  try {
    // PUT sets the quantity, so read it fresh first (staleTime 0): a number
    // cached before another tab changed the cart would undo that change.
    const cart = await client.fetchQuery({ queryKey: SERVER_CART_KEY, queryFn: () => getCart(), staleTime: 0 });
    const line = cart.items.find((l) => l.offerId === offerId);
    const quantity = Math.min((line?.quantity ?? 0) + 1, MAX_QTY);
    client.setQueryData(SERVER_CART_KEY, await upsertItem(offerId, quantity));
    return "added";
  } catch (e) {
    // The server refuses a 21st line with 422 (API.md, Cart).
    if (e instanceof ApiError && e.status === 422) return "full";
    return "error";
  }
}
