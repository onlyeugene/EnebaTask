"use client";

import { ChevronDown } from "lucide-react";
import { ErrorState } from "@/components/ui/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { useFilters } from "../hooks/use-filters";
import { useStoreParams } from "../hooks/use-store-params";
import { CheckList } from "./check-list";
import { FilterSection } from "./filter-section";
import { PriceRange } from "./price-range";

// The options and counts come from GET /api/products/filters. Counts are for
// the whole in-stock store, not the current selection (see PLAN.md).
// Its own query: if it fails, the product grid still works.
export function FilterPanel() {
  const filters = useFilters();
  const { params, list, update, toggle } = useStoreParams();

  if (filters.isPending) return <Skeleton className="h-150 bg-brand-mid" />;
  if (filters.isError) {
    return (
      <ErrorState
        message="Filters couldn't load."
        onRetry={() => void filters.refetch()}
        className="px-3 py-6"
      />
    );
  }

  const f = filters.data;
  const minPrice = params.get("minPrice");
  const maxPrice = params.get("maxPrice");
  const region = list("region");

  return (
    <aside aria-label="Filters" className="bg-brand-mid">
      {/* Eneba says "Price range (EUR)"; it wraps in our narrower 192px panel. */}
      <FilterSection title="Price (EUR)" active={minPrice || maxPrice ? 1 : 0}>
        {/* The key remounts the inputs when the URL prices change from
            outside (a chip, "Clear all", Back), so they show the new values. */}
        <PriceRange
          key={`${minPrice}-${maxPrice}`}
          minCents={minPrice}
          maxCents={maxPrice}
          bounds={f.price}
          onApply={(min, max) => update({ minPrice: min, maxPrice: max })}
        />
      </FilterSection>

      {/* One region at a time, like Eneba's "Country" select. */}
      <FilterSection title="Region" active={region.length}>
        <div className="relative">
          <select
            aria-label="Region"
            value={region[0] ?? ""}
            onChange={(e) => update({ region: e.target.value || null })}
            className="h-11 w-full cursor-pointer appearance-none border border-brand-border bg-brand pr-9 pl-3 text-base text-text-secondary [&>option]:bg-brand-darkest"
          >
            <option value="">All regions</option>
            {f.regions.map((o) => (
              <option key={o.value} value={o.value}>
                {o.label} ({o.count})
              </option>
            ))}
          </select>
          <ChevronDown className="pointer-events-none absolute top-3 right-3 size-5" aria-hidden />
        </div>
      </FilterSection>

      <FilterSection title="Type" active={list("type").length}>
        <CheckList options={f.types} selected={list("type")} onToggle={(v) => toggle("type", v)} />
      </FilterSection>

      <FilterSection title="Platform" active={list("platform").length}>
        <CheckList
          options={f.platforms}
          selected={list("platform")}
          onToggle={(v) => toggle("platform", v)}
        />
      </FilterSection>

      <FilterSection title="Genre" active={list("genre").length}>
        <CheckList
          options={f.genres}
          selected={list("genre")}
          onToggle={(v) => toggle("genre", v)}
          limit={6}
        />
      </FilterSection>
    </aside>
  );
}
