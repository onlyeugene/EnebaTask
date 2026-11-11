import { useEffect } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { getTokens } from "@/lib/tokens";
import { mergeGuestCart } from "../merge";
import { useGuestCart } from "../store";

// Reads the saved guest cart after the first render (see skipHydration in
// store.ts), and again whenever another tab changes it. Without the second
// part, this tab would keep its old list and overwrite the other tab's
// additions on its next change. Mounted by the shop header's cart popover,
// and by the cart page, whose checkout layout has no popover.
//
// If a signed-in page still finds guest lines, the merge at login failed
// (the guest cart is only cleared after a merge succeeds), so try it again
// here. A failure stays quiet: the lines are kept for the next page.
export function useGuestCartSync() {
  const client = useQueryClient();

  useEffect(() => {
    void useGuestCart.persist.rehydrate()?.then(() => {
      if (getTokens() && useGuestCart.getState().items.length > 0) {
        mergeGuestCart(client).catch(() => {});
      }
    });

    const onStorage = (e: StorageEvent) => {
      if (e.key === useGuestCart.persist.getOptions().name) void useGuestCart.persist.rehydrate();
    };
    window.addEventListener("storage", onStorage);
    return () => window.removeEventListener("storage", onStorage);
  }, [client]);
}
