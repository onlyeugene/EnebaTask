"use client";

import type { ReactNode } from "react";
import type { UseQueryResult } from "@tanstack/react-query";
import type { ProductList } from "@/types/api";
import { ApiError } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { ErrorState } from "@/components/ui/error-state";
import { ProductCard } from "@/components/ui/product-card";
import { CardSkeleton } from "@/components/ui/skeleton";
import { useStoreParams } from "../hooks/use-store-params";

// The grid with its loading, error and empty states.
export function ProductResults({ query }: { query: UseQueryResult<ProductList> }) {
  const { q, clear } = useStoreParams();

  // As on the home page: while a Retry runs, show the skeleton, not the old error.
  if (query.isPending || (query.isError && query.isFetching)) {
    return (
      <Grid busy>
        {Array.from({ length: 8 }, (_, i) => (
          <CardSkeleton key={i} />
        ))}
      </Grid>
    );
  }

  if (query.isError) {
    // A 400 means a bad value in the link (old or edited by hand). Retrying
    // gives the same answer, so offer to clear the filters instead.
    if (query.error instanceof ApiError && query.error.status === 400) {
      return (
        <ErrorState
          message={`This link has a filter we can't use (${query.error.message}).`}
          onRetry={clear}
          retryLabel="Clear filters"
        />
      );
    }
    return (
      <ErrorState message="Products couldn't load." onRetry={() => void query.refetch()} />
    );
  }

  if (query.data.items.length === 0) {
    return (
      <div className="border border-divider bg-brand-darkest px-6 py-12 text-center">
        <p className="text-xl font-extrabold">No products found</p>
        <p className="mt-3 text-base text-text-secondary">
          {q ? "Check the spelling, or try fewer filters." : "Try fewer filters."}
        </p>
        <Button variant="outline" className="mt-6" onClick={clear}>
          Clear all filters
        </Button>
      </div>
    );
  }

  // isPlaceholderData: the old page is still shown while the new one loads.
  return (
    <Grid busy={query.isPlaceholderData}>
      {query.data.items.map((p) => (
        <ProductCard key={p.id} product={p} />
      ))}
    </Grid>
  );
}

// Fixed 224px columns (the card's width): 4 on desktop next to the filters,
// fewer as the screen narrows.
function Grid({ busy, children }: { busy: boolean; children: ReactNode }) {
  return (
    <div
      aria-busy={busy}
      className={`grid grid-cols-[repeat(auto-fill,224px)] justify-center gap-5 transition-opacity sm:justify-start ${busy ? "opacity-60" : ""}`}
    >
      {children}
    </div>
  );
}
