"use client";

import { useEffect, useState } from "react";
import { QueryClientProvider, type QueryClient } from "@tanstack/react-query";
import { makeQueryClient } from "./query-client";
import { TOKENS_KEY } from "./tokens";

export function Providers({ children }: { children: React.ReactNode }) {
  // useState keeps one client for the life of the tab. A module-level client
  // would also be shared between requests during server rendering.
  const [client] = useState(makeQueryClient);
  useOtherTabAuth(client);
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}

// When another tab logs in or out, the "storage" event fires here. Reload who
// is signed in, the cart and the wishlist, so this tab switches between the guest cart and
// the account cart without a page reload. Only a change between "no tokens"
// and "tokens" counts: a token refresh in another tab changes the value too,
// and needs nothing here. The keys are spelled out because lib can't import
// features: ["me"] is features/auth's ME_KEY, ["cart"] covers both carts,
// ["wishlist"] is features/wishlist's WISHLIST_KEY.
function useOtherTabAuth(client: QueryClient) {
  useEffect(() => {
    const onStorage = (e: StorageEvent) => {
      if (e.key !== TOKENS_KEY || (e.oldValue === null) === (e.newValue === null)) return;
      void client.invalidateQueries({ queryKey: ["me"] });
      void client.invalidateQueries({ queryKey: ["cart"] });
      void client.invalidateQueries({ queryKey: ["wishlist"] });
    };
    window.addEventListener("storage", onStorage);
    return () => window.removeEventListener("storage", onStorage);
  }, [client]);
}
