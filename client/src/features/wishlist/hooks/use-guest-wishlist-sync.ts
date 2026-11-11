import { useEffect } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { getTokens } from "@/lib/tokens";
import { mergeGuestWishlist } from "../merge";
import { useGuestWishlist } from "../store";

// The cart's useGuestCartSync, for the wishlist: reads the saved guest list
// after mount and whenever another tab changes it, and retries a merge that
// failed at login. Mounted once, by the header's wishlist link.
export function useGuestWishlistSync() {
  const client = useQueryClient();

  useEffect(() => {
    void useGuestWishlist.persist.rehydrate()?.then(() => {
      if (getTokens() && useGuestWishlist.getState().items.length > 0) {
        mergeGuestWishlist(client).catch(() => {});
      }
    });

    const onStorage = (e: StorageEvent) => {
      if (e.key === useGuestWishlist.persist.getOptions().name) void useGuestWishlist.persist.rehydrate();
    };
    window.addEventListener("storage", onStorage);
    return () => window.removeEventListener("storage", onStorage);
  }, [client]);
}
