"use client";

import { ArrowDownWideNarrow, ChevronDown } from "lucide-react";
import { useStoreParams } from "../hooks/use-store-params";

// Values are the API's `sort` param; labels follow Eneba's "Popularity: Most popular".
const SORTS = [
  { value: "relevance", label: "Relevance: Best match" },
  { value: "bestselling", label: "Popularity: Most popular" },
  { value: "price_asc", label: "Price: Low to high" },
  { value: "price_desc", label: "Price: High to low" },
  { value: "discount", label: "Discount: Biggest first" },
  { value: "newest", label: "Release: Newest first" },
];

// A native <select> styled as bold text: keyboard and screen reader support
// come for free, and the open list is the system's own.
export function SortSelect() {
  const { params, q, update } = useStoreParams();
  // Same defaults as the API. Picking the default removes `sort` from the URL.
  const fallback = q ? "relevance" : "bestselling";
  const value = params.get("sort") ?? fallback;
  // Relevance only means something with a search text.
  const options = q ? SORTS : SORTS.filter((s) => s.value !== "relevance");

  return (
    <label className="relative flex items-center gap-2 text-base font-bold">
      <ArrowDownWideNarrow className="size-5" aria-hidden />
      <span className="sr-only">Sort by</span>
      <select
        value={value}
        onChange={(e) => update({ sort: e.target.value === fallback ? null : e.target.value })}
        // field-sizing: the box fits the picked label, so the chevron sits
        // right after it. Browsers without it fall back to the widest label.
        className="cursor-pointer appearance-none bg-transparent pr-6 font-bold [field-sizing:content] [&>option]:bg-brand-darkest"
      >
        {options.map((s) => (
          <option key={s.value} value={s.value}>
            {s.label}
          </option>
        ))}
      </select>
      <ChevronDown className="pointer-events-none absolute right-0 size-5" aria-hidden />
    </label>
  );
}
