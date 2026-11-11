"use client";

import { X } from "lucide-react";
import { useStoreParams, type ListFilter } from "../hooks/use-store-params";

const GROUPS: { name: ListFilter; label: string }[] = [
  { name: "type", label: "Type" },
  { name: "platform", label: "Platform" },
  { name: "region", label: "Region" },
  { name: "genre", label: "Genre" },
];

const CHIP =
  "flex h-10 items-center gap-2 bg-brand-darkest px-4 text-base font-semibold transition-colors hover:bg-brand-dark";

// The chips above the grid in store2.png: "Type (1)", "Regions (5)", "Clear all".
// On Eneba a chip opens a dropdown; here it removes that whole filter, since
// the same options are already in the panel on the left.
export function ActiveFilters() {
  const { params, list, update, clear } = useStoreParams();
  const chips = GROUPS.map((g) => ({ ...g, count: list(g.name).length })).filter((g) => g.count > 0);
  const hasPrice = params.has("minPrice") || params.has("maxPrice");
  if (chips.length === 0 && !hasPrice) return null;

  return (
    <div className="mb-5 flex flex-wrap items-center gap-3">
      {chips.map((c) => (
        <button
          key={c.name}
          type="button"
          onClick={() => update({ [c.name]: null })}
          aria-label={`Remove ${c.label} filter`}
          className={CHIP}
        >
          {c.label} ({c.count})
          <X className="size-4" aria-hidden />
        </button>
      ))}
      {hasPrice && (
        <button
          type="button"
          onClick={() => update({ minPrice: null, maxPrice: null })}
          aria-label="Remove price filter"
          className={CHIP}
        >
          Price
          <X className="size-4" aria-hidden />
        </button>
      )}
      <button type="button" onClick={clear} className="link-hover px-2 text-base text-yellow">
        Clear all
      </button>
    </div>
  );
}
