"use client";

import { ArrowDownWideNarrow, ChevronDown } from "lucide-react";
import type { WishlistEntry } from "@/types/api";

export const SORTS = [
  { value: "newest", label: "Date: Newest first" },
  { value: "oldest", label: "Date: Oldest first" },
  { value: "price_asc", label: "Price: Low to high" },
] as const;

export type WishlistSortValue = (typeof SORTS)[number]["value"];

// Sorted in the browser: the whole list is already here. Sold out products
// have no price, so the price sort puts them last.
export function sortEntries(entries: WishlistEntry[], sort: WishlistSortValue) {
  const time = (e: WishlistEntry) => Date.parse(e.addedAt);
  const price = (e: WishlistEntry) => e.product.minPriceCents ?? Number.MAX_SAFE_INTEGER;
  const by = {
    newest: (a: WishlistEntry, b: WishlistEntry) => time(b) - time(a),
    oldest: (a: WishlistEntry, b: WishlistEntry) => time(a) - time(b),
    price_asc: (a: WishlistEntry, b: WishlistEntry) => price(a) - price(b),
  }[sort];
  return [...entries].sort(by);
}

type Props = { value: WishlistSortValue; onChange: (v: WishlistSortValue) => void };

// Same look as the store's sort (store/components/sort-select.tsx), but kept
// in component state: the list is per user, so a shared link would mean nothing.
export function WishlistSort({ value, onChange }: Props) {
  return (
    <label className="relative flex items-center gap-2 text-base font-bold">
      <ArrowDownWideNarrow className="size-5" aria-hidden />
      <span className="sr-only">Sort by</span>
      <select
        value={value}
        onChange={(e) => onChange(e.target.value as WishlistSortValue)}
        className="cursor-pointer appearance-none bg-transparent pr-6 font-bold [field-sizing:content] [&>option]:bg-brand-darkest"
      >
        {SORTS.map((s) => (
          <option key={s.value} value={s.value}>
            {s.label}
          </option>
        ))}
      </select>
      <ChevronDown className="pointer-events-none absolute right-0 size-5" aria-hidden />
    </label>
  );
}
