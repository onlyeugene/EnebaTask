import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { quoteCart } from "../api";
import { useGuestCart } from "../store";
import { EMPTY_CART, type CartView } from "./cart-view";

// The guest cart (localStorage) with live prices from POST /api/cart/quote.
// `enabled` is false while signed in, so no quote is asked for then.
export function useGuestCartView(enabled: boolean): CartView {
  const items = useGuestCart((s) => s.items);
  const hydrated = useGuestCart((s) => s.hydrated);
  const { step, remove, clear } = useGuestCart.getState();

  const query = useQuery({
    // The items are part of the key, so every change asks for a new quote.
    queryKey: ["cart", "quote", items],
    queryFn: ({ signal }) => quoteCart(items, signal),
    enabled: enabled && hydrated && items.length > 0,
    // Keep the old cart on screen (dimmed) while the new quote loads.
    placeholderData: keepPreviousData,
  });

  const empty = hydrated && items.length === 0;
  return {
    cart: empty ? EMPTY_CART : query.data,
    // Badge count from local state, so it updates at once with no request.
    itemCount: items.reduce((n, i) => n + i.quantity, 0),
    loading: !hydrated || (!empty && query.isPending),
    busy: !empty && query.isPlaceholderData,
    error: empty ? null : query.error,
    retry: () => void query.refetch(),
    step,
    remove,
    clear,
  };
}
