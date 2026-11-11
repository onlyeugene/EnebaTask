"use client";

import { useState } from "react";
import { QueryClientProvider } from "@tanstack/react-query";
import { makeQueryClient } from "./query-client";

export function Providers({ children }: { children: React.ReactNode }) {
  // useState keeps one client for the life of the tab. A module-level client
  // would also be shared between requests during server rendering.
  const [client] = useState(makeQueryClient);
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}
