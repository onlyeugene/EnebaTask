import type { QueryClient } from "@tanstack/react-query";
import { mergeWishlist, WISHLIST_KEY } from "./api";
import { useGuestWishlist } from "./store";

// Moves the guest wishlist into the account at login: the server keeps the
// union, each product once. Like the cart merge (cart/merge.ts), the guest
// list is only cleared after the server answers 200, so a failed request
// loses nothing, and sending it again is safe (a product already saved is
// left as it is).
export async function mergeGuestWishlist(client: QueryClient) {
  await useGuestWishlist.persist.rehydrate();
  const sent = useGuestWishlist.getState().items;
  if (sent.length === 0) return;

  const wishlist = await mergeWishlist(sent.map(({ productId, addedAt }) => ({ productId, addedAt })));
  client.setQueryData(WISHLIST_KEY, wishlist);

  // Remove only what was sent: an item saved in another tab meanwhile stays for the next merge.
  useGuestWishlist.setState((s) => ({
    items: s.items.filter((i) => !sent.some((x) => x.productId === i.productId)),
  }));
}
