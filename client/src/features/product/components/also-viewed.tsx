"use client";

import { useState } from "react";
import type { ProductDetail } from "@/types/api";
import { Button } from "@/components/ui/button";
import { ProductCard } from "@/components/ui/product-card";
import { useWishlist } from "@/features/wishlist/hooks/use-wishlist";
import { useAddCheapest } from "../hooks/use-add-cheapest";
import { useAlsoViewed } from "../hooks/use-also-viewed";

// Eneba's "Gamers also viewed". We have no view history, so it shows other
// products of the same type and platform. 5 at first, "Load more" adds 5.
// The list endpoint caps pageSize at 60, so it stops before that.
const STEP = 5;
const MAX = 55;

export function AlsoViewed({ product }: { product: ProductDetail }) {
  const [limit, setLimit] = useState(STEP);
  const { items, hasMore, loading } = useAlsoViewed(product, limit);
  const addCheapest = useAddCheapest();
  const wishlist = useWishlist();

  // Nothing to show (or it failed): leave the section out, the page works without it.
  if (items.length === 0) return null;

  return (
    <section>
      <h2 className="mb-6 text-xl font-extrabold">Gamers also viewed</h2>
      <div className="grid grid-cols-[repeat(auto-fill,224px)] justify-center gap-5 sm:justify-start">
        {items.map((p) => (
          <ProductCard
                  key={p.id}
                  product={p}
                  onAddToCart={addCheapest}
                  wishlisted={wishlist.has(p.id)}
                  onToggleWishlist={wishlist.toggle}
                />
        ))}
      </div>
      {hasMore && limit < MAX && (
        <div className="mt-8 flex justify-center">
          <Button
            variant="outline"
            className="w-64 text-base"
            disabled={loading}
            onClick={() => setLimit((l) => l + STEP)}
          >
            Load more
          </Button>
        </div>
      )}
    </section>
  );
}
