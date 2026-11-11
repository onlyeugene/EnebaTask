import { QueryClient } from "@tanstack/react-query";
import { ApiError } from "./client";

export function makeQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 30_000,
        refetchOnWindowFocus: false,
        // A 4xx (404, 422, ...) gives the same answer on retry, so show it at
        // once. Network errors and 5xx get two more tries.
        retry: (failures, error) => {
          if (error instanceof ApiError && error.status < 500) return false;
          return failures < 2;
        },
      },
    },
  });
}
