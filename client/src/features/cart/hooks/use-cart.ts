import { useMe } from "@/features/auth/hooks/use-me";
import type { CartView } from "./cart-view";
import { useGuestCartView } from "./use-guest-cart-view";
import { useServerCart } from "./use-server-cart";

// The cart on screen: the server cart when signed in, the guest cart
// otherwise. Components never need to know which one they show.
// Both hooks always run (hooks can't be called conditionally); `enabled`
// stops the unused one from sending requests.
export function useCart(): CartView {
  const me = useMe();
  const signedIn = !!me.data;
  const guest = useGuestCartView(me.isSuccess && !signedIn);
  const server = useServerCart(signedIn);

  // Until /me answers we don't know which cart to show.
  if (me.isPending) return { ...guest, cart: undefined, itemCount: 0, loading: true, error: null };
  // Showing the guest cart here would hide a signed-in user's real cart.
  if (me.isError) return { ...guest, cart: undefined, loading: false, error: me.error, retry: () => void me.refetch() };
  return signedIn ? server : guest;
}
