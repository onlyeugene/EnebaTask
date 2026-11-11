import { keepPreviousData, useQuery } from "@tanstack/react-query";
import type { Cart } from "@/types/api";
import { quoteCart } from "../api";
import { useGuestCart } from "../store";

const EMPTY: Cart = { items: [], itemCount: 0, totalCents: 0, cashbackCents: 0 };

// The guest cart with live prices. Phase 14 adds the signed-in branch here
// (GET /api/cart), so components never need to know which cart they show.
export function useCart() {
  const items = useGuestCart((s) => s.items);
  const hydrated = useGuestCart((s) => s.hydrated);
  const { step, remove, clear } = useGuestCart.getState();

  const query = useQuery({
    // The items are part of the key, so every change asks for a new quote.
    queryKey: ["cart", "quote", items],
    queryFn: ({ signal }) => quoteCart(items, signal),
    enabled: hydrated && items.length > 0,
    // Keep the old cart on screen (dimmed) while the new quote loads.
    placeholderData: keepPreviousData,
  });

  const empty = hydrated && items.length === 0;
  return {
    cart: empty ? EMPTY : query.data,
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
