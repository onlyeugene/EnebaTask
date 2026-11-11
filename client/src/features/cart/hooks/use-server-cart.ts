import { useIsMutating, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { Cart } from "@/types/api";
import { clearHeldOrder } from "@/features/checkout/held-order";
import { getCart, removeItem, SERVER_CART_KEY, upsertItem } from "../api";
import { MAX_QTY } from "../store";
import type { CartView } from "./cart-view";

const WRITE_KEY = ["cart", "write"];

// The signed-in user's cart: GET /api/cart, changed with PUT and DELETE.
//
// A click changes the cached cart at once (so quick +/+ clicks each count from
// the number on screen), then sends the new quantity. `scope` makes TanStack
// send the writes one at a time in click order: PUT sets a quantity, so if
// "3" overtook "4" on the way, the cart would end on 3. Only the last write's
// answer is put on screen, so an earlier answer can't briefly undo a later click.
export function useServerCart(enabled: boolean): CartView {
  const client = useQueryClient();
  const query = useQuery({
    queryKey: SERVER_CART_KEY,
    queryFn: ({ signal }) => getCart(signal),
    enabled,
  });
  const writing = useIsMutating({ mutationKey: WRITE_KEY }) > 0;

  const write = useMutation({
    mutationKey: WRITE_KEY,
    scope: { id: "server-cart" },
    mutationFn: ({ offerId, quantity }: { offerId: string; quantity: number | null }) =>
      quantity === null ? removeItem(offerId) : upsertItem(offerId, quantity),
    onSettled: (cart, error) => {
      if (client.isMutating({ mutationKey: WRITE_KEY }) > 1) return; // more writes queued
      if (cart && !error) client.setQueryData(SERVER_CART_KEY, cart);
      // On failure, reload the real cart, which drops the change made on screen.
      else void client.invalidateQueries({ queryKey: SERVER_CART_KEY });
    },
  });

  // null quantity = remove the line.
  const change = (offerId: string, quantity: number | null) => {
    // An order held for checkout was reserved from the old cart; Pay must
    // reserve a new one that matches this cart (checkout/held-order.ts).
    clearHeldOrder();
    // A GET still on its way would overwrite the change with the old cart.
    void client.cancelQueries({ queryKey: SERVER_CART_KEY });
    client.setQueryData<Cart>(SERVER_CART_KEY, (c) =>
      c && {
        ...c,
        items:
          quantity === null
            ? c.items.filter((l) => l.offerId !== offerId)
            : c.items.map((l) => (l.offerId === offerId ? { ...l, quantity } : l)),
      },
    );
    write.mutate({ offerId, quantity });
  };

  const step = (offerId: string, delta: number) => {
    const line = client.getQueryData<Cart>(SERVER_CART_KEY)?.items.find((l) => l.offerId === offerId);
    if (!line) return;
    const quantity = Math.min(Math.max(line.quantity + delta, 1), MAX_QTY);
    if (quantity !== line.quantity) change(offerId, quantity);
  };

  const cart = query.data;
  return {
    cart,
    // Summed here rather than read from cart.itemCount, which a click doesn't update.
    itemCount: cart?.items.reduce((n, l) => n + l.quantity, 0) ?? 0,
    loading: query.isPending,
    busy: writing || (query.isFetching && !query.isPending),
    error: query.error,
    retry: () => void query.refetch(),
    step,
    remove: (offerId) => change(offerId, null),
  };
}
