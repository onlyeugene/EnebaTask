import { useEffect } from "react";
import { useGuestCart } from "../store";

// Reads the saved guest cart after the first render (see skipHydration in
// store.ts), and again whenever another tab changes it. Without the second
// part, this tab would keep its old list and overwrite the other tab's
// additions on its next change. Mounted by the shop header's cart popover,
// and by the cart page, whose checkout layout has no popover.
export function useGuestCartSync() {
  useEffect(() => {
    void useGuestCart.persist.rehydrate();

    const onStorage = (e: StorageEvent) => {
      if (e.key === useGuestCart.persist.getOptions().name) void useGuestCart.persist.rehydrate();
    };
    window.addEventListener("storage", onStorage);
    return () => window.removeEventListener("storage", onStorage);
  }, []);
}
