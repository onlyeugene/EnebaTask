import { useMemo } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { ProductCard, Wishlist, WishlistEntry } from "@/types/api";
import { useMe } from "@/features/auth/hooks/use-me";
import { addToWishlist, getWishlist, removeFromWishlist, WISHLIST_KEY } from "../api";
import { useGuestWishlist } from "../store";

const WRITE_KEY = ["wishlist", "write"];

type Write = { product: ProductCard; add: boolean; previous?: WishlistEntry };

// Which products are wishlisted, and a toggle for the hearts. Like useCart, it
// reads the server list when signed in and the guest list otherwise.
//
// Signed in, a click changes the cached list at once, then sends PUT or
// DELETE. `scope` sends the writes one at a time in click order, so quick
// on/off clicks end on the last one. If a write fails, that one change is
// undone on screen and the list is reloaded.
export function useWishlist() {
  const client = useQueryClient();
  const me = useMe();
  const signedIn = !!me.data;
  const query = useQuery({
    queryKey: WISHLIST_KEY,
    queryFn: ({ signal }) => getWishlist(signal),
    enabled: signedIn,
  });
  const guest = useGuestWishlist();

  const ids = useMemo(() => {
    if (signedIn) return new Set(query.data?.items.map((e) => e.product.id));
    // Until /me answers we don't know which list to show, so show none.
    return new Set(me.isSuccess ? guest.items.map((i) => i.productId) : []);
  }, [signedIn, me.isSuccess, query.data, guest.items]);

  const write = useMutation({
    mutationKey: WRITE_KEY,
    scope: { id: "wishlist" },
    mutationFn: ({ product, add }: Write) => (add ? addToWishlist(product.id) : removeFromWishlist(product.id)),
    onError: (_err, { product, add, previous }) => {
      client.setQueryData<Wishlist>(WISHLIST_KEY, (w) => w && (add ? without(w, product.id) : previous ? withEntry(w, previous) : w));
    },
    onSettled: (list, error) => {
      if (client.isMutating({ mutationKey: WRITE_KEY }) > 1) return; // more writes queued
      if (list && !error) client.setQueryData(WISHLIST_KEY, list);
      else void client.invalidateQueries({ queryKey: WISHLIST_KEY });
    },
  });

  const toggle = (product: ProductCard) => {
    if (!me.isSuccess) return;
    const add = !ids.has(product.id);
    if (!signedIn) {
      if (add) guest.add(product.id, product.slug);
      else guest.remove(product.id);
      return;
    }
    // A GET still on its way would overwrite the change with the old list.
    void client.cancelQueries({ queryKey: WISHLIST_KEY });
    const previous = client.getQueryData<Wishlist>(WISHLIST_KEY)?.items.find((e) => e.product.id === product.id);
    client.setQueryData<Wishlist>(WISHLIST_KEY, (w) =>
      w && (add ? withEntry(w, { product, addedAt: new Date().toISOString() }) : without(w, product.id)),
    );
    write.mutate({ product, add, previous });
  };

  return {
    has: (productId: string) => ids.has(productId),
    count: ids.size,
    toggle,
  };
}

function without(w: Wishlist, productId: string): Wishlist {
  return { items: w.items.filter((e) => e.product.id !== productId) };
}

function withEntry(w: Wishlist, entry: WishlistEntry): Wishlist {
  return { items: [...without(w, entry.product.id).items, entry] };
}
