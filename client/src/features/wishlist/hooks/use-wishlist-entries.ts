import { useQueries, useQuery } from "@tanstack/react-query";
import type { WishlistEntry } from "@/types/api";
import { ApiError } from "@/lib/client";
import { useMe } from "@/features/auth/hooks/use-me";
import { productQuery } from "@/features/product/hooks/use-product";
import { getWishlist, WISHLIST_KEY } from "../api";
import { useGuestWishlist } from "../store";

type Result = { entries: WishlistEntry[]; loading: boolean; error: unknown; retry: () => void };

// The wishlist page's products with their dates.
// Signed in: GET /api/wishlist already returns the cards.
// Guest: the browser only has ids and slugs, so each product is loaded with
// GET /api/products/{slug} (the product page's query, so they share a cache).
// One request per product is fine for a list capped at 100, and it uses no
// new endpoint. A product that is gone (404) is left out.
export function useWishlistEntries(): Result {
  const me = useMe();
  const signedIn = !!me.data;
  const server = useQuery({
    queryKey: WISHLIST_KEY,
    queryFn: ({ signal }) => getWishlist(signal),
    enabled: signedIn,
  });
  const guestItems = useGuestWishlist((s) => s.items);
  const hydrated = useGuestWishlist((s) => s.hydrated);
  const asGuest = me.isSuccess && !signedIn;
  const guest = useQueries({ queries: (asGuest ? guestItems : []).map((i) => productQuery(i.slug)) });

  if (me.isPending) return { entries: [], loading: true, error: null, retry: () => {} };
  if (me.isError) return { entries: [], loading: false, error: me.error, retry: () => void me.refetch() };

  if (signedIn) {
    return {
      entries: server.data?.items ?? [],
      loading: server.isPending,
      error: server.error,
      retry: () => void server.refetch(),
    };
  }

  const failed = guest.find((q) => q.isError && !(q.error instanceof ApiError && q.error.status === 404));
  const entries: WishlistEntry[] = [];
  guestItems.forEach((item, i) => {
    const product = guest[i]?.data?.product;
    if (product) entries.push({ product, addedAt: item.addedAt });
  });
  return {
    entries,
    loading: !hydrated || guest.some((q) => q.isPending),
    error: failed?.error ?? null,
    retry: () => guest.forEach((q) => q.isError && void q.refetch()),
  };
}
