"use client";

import { useState } from "react";
import Link from "next/link";
import { ChevronRight, SlidersHorizontal } from "lucide-react";
import { useProducts } from "../hooks/use-products";
import { useStoreParams } from "../hooks/use-store-params";
import { ActiveFilters } from "./active-filters";
import { FilterPanel } from "./filter-panel";
import { Pagination } from "./pagination";
import { ProductResults } from "./product-results";
import { SortSelect } from "./sort-select";

// Layout from store.png: filters on the left, chips, count and sort on top,
// a grid of product cards below. 192px sidebar + 20px gap + 4 cards of 224px
// with 20px gaps = 1168px, the page width inside its 16px side padding.
export function StoreContent() {
  const { params, q } = useStoreParams();
  const products = useProducts(params.toString());
  // Phones only: the filter panel is folded away behind a button.
  const [showFilters, setShowFilters] = useState(false);

  return (
    <>
      <nav aria-label="Breadcrumb" className="mb-4 flex items-center gap-1 text-sm text-text-secondary">
        <Link href="/" className="link-hover">
          Home
        </Link>
        <ChevronRight className="size-4" aria-hidden />
        <span>Store</span>
      </nav>
      <h1 className="mb-8 text-3xl font-extrabold">
        {q ? `Results for “${q}”` : "Games, Gift Cards, Software & More | Best Deals"}
      </h1>

      <div className="grid gap-5 md:grid-cols-[192px_1fr]">
        <div>
          <button
            type="button"
            aria-expanded={showFilters}
            onClick={() => setShowFilters((s) => !s)}
            className="flex h-11 w-full items-center justify-center gap-2 bg-brand-mid text-base font-bold md:hidden"
          >
            <SlidersHorizontal className="size-5" aria-hidden />
            {showFilters ? "Hide filters" : "Filters"}
          </button>
          <div className={`${showFilters ? "mt-3 block" : "hidden"} md:mt-0 md:block`}>
            <FilterPanel />
          </div>
        </div>

        <div className="min-w-0">
          <ActiveFilters />
          <div className="mb-5 flex flex-wrap items-center justify-between gap-3">
            <p className="text-base">
              Results found:{" "}
              <strong className="font-extrabold">{products.data?.total ?? "…"}</strong>
            </p>
            <SortSelect />
          </div>
          <ProductResults query={products} />
          {products.data && (
            <Pagination
              page={products.data.page}
              pageSize={products.data.pageSize}
              total={products.data.total}
            />
          )}
        </div>
      </div>
    </>
  );
}
