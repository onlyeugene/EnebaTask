"use client";

import { useState, type ReactNode } from "react";
import type { FilterOption, ProductType, WishlistEntry } from "@/types/api";
import { typeLabel } from "@/lib/format";
import { ErrorState } from "@/components/ui/error-state";
import { ProductCard } from "@/components/ui/product-card";
import { CardSkeleton } from "@/components/ui/skeleton";
import { useAddCheapest } from "@/features/product/hooks/use-add-cheapest";
import { useWishlist } from "../hooks/use-wishlist";
import { useWishlistEntries } from "../hooks/use-wishlist-entries";
import { WishlistEmpty } from "./wishlist-empty";
import { WishlistFilters } from "./wishlist-filters";
import { sortEntries, WishlistSort, type WishlistSortValue } from "./wishlist-sort";

const TYPE_ORDER: ProductType[] = ["game", "gift_card", "software"];

// /wishlist (favorites.png). Left out from Eneba: the email price alert
// banner, since we send no emails. Same 192px panel + 224px card grid as the store.
// pb-10: each page ends with its own 40px before the footer.
export function WishlistView() {
  const { entries, loading, error, retry } = useWishlistEntries();
  const wishlist = useWishlist();
  const addCheapest = useAddCheapest();
  const [name, setName] = useState("");
  const [types, setTypes] = useState<string[]>([]);
  const [sort, setSort] = useState<WishlistSortValue>("newest");

  const words = name.trim().toLowerCase();
  const shown = sortEntries(
    entries.filter(
      (e) =>
        (types.length === 0 || types.includes(e.product.type)) &&
        e.product.name.toLowerCase().includes(words),
    ),
    sort,
  );

  let body: ReactNode;
  if (loading) {
    body = (
      <Grid>
        {Array.from({ length: 4 }, (_, i) => (
          <CardSkeleton key={i} />
        ))}
      </Grid>
    );
  } else if (error) {
    body = <ErrorState message="Your wishlist couldn't load." onRetry={retry} />;
  } else if (entries.length === 0) {
    body = <WishlistEmpty />;
  } else {
    body = (
      <>
        <div className="mb-5 flex flex-wrap items-center justify-between gap-3">
          <p className="text-base">
            Results found: <strong className="font-extrabold">{shown.length}</strong>
          </p>
          <WishlistSort value={sort} onChange={setSort} />
        </div>
        {shown.length === 0 ? (
          <p className="py-12 text-center text-base text-text-secondary">No products match these filters.</p>
        ) : (
          <Grid>
            {shown.map((e) => (
              <ProductCard
                key={e.product.id}
                product={e.product}
                onAddToCart={addCheapest}
                wishlisted={wishlist.has(e.product.id)}
                onToggleWishlist={wishlist.toggle}
              />
            ))}
          </Grid>
        )}
      </>
    );
  }

  return (
    <div className="mx-auto max-w-page px-4 pt-6 pb-10">
      <h1 className="mb-8 text-3xl font-extrabold">Wishlist</h1>
      <div className="grid gap-5 md:grid-cols-[192px_1fr]">
        <WishlistFilters
          name={name}
          onName={setName}
          types={typeOptions(entries)}
          selectedTypes={types}
          onToggleType={(t) => setTypes((ts) => (ts.includes(t) ? ts.filter((x) => x !== t) : [...ts, t]))}
        />
        <div className="min-w-0">{body}</div>
      </div>
    </div>
  );
}

// Counts are for the whole list, like the store's (they don't change as you tick).
function typeOptions(entries: WishlistEntry[]): FilterOption[] {
  return TYPE_ORDER.map((t) => ({
    value: t,
    label: typeLabel(t),
    count: entries.filter((e) => e.product.type === t).length,
  })).filter((o) => o.count > 0);
}

function Grid({ children }: { children: ReactNode }) {
  return (
    <div className="grid grid-cols-[repeat(auto-fill,224px)] justify-center gap-5 sm:justify-start">
      {children}
    </div>
  );
}
