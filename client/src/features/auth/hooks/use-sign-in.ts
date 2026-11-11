import { useRouter } from "next/navigation";
import { useQueryClient } from "@tanstack/react-query";
import type { AuthResponse } from "@/types/api";
import { setTokens } from "@/lib/tokens";
import { mergeGuestCart } from "@/features/cart/merge";
import { mergeGuestWishlist } from "@/features/wishlist/merge";
import { ME_KEY } from "./use-me";

// The steps after a successful login or register, shared by both forms:
// save the tokens, then move the guest cart (PLAN.md 5b) and wishlist into the account,
// then open the page the user came from.
// The merge is awaited, so the cart is already merged when that page shows.
// If it fails the user is still signed in and the guest cart is kept (it is
// only cleared after a merge succeeds); useGuestCartSync tries again on the
// next page that shows the cart.
export function useSignIn() {
  const router = useRouter();
  const client = useQueryClient();

  return async (res: AuthResponse, next: string) => {
    setTokens(res);
    // Drop any cart cached for a previous account before loading this one's.
    client.removeQueries({ queryKey: ["cart"] });
    client.removeQueries({ queryKey: ["wishlist"] });
    client.setQueryData(ME_KEY, res.user);
    // The wishlist merge follows the same rule: guest list cleared only on success.
    await Promise.all([mergeGuestCart(client).catch(() => {}), mergeGuestWishlist(client).catch(() => {})]);
    router.replace(next);
  };
}
