"use client";

import { Search } from "lucide-react";
import type { FilterOption } from "@/types/api";
import { CheckList } from "@/features/store/components/check-list";
import { FilterSection } from "@/features/store/components/filter-section";

type Props = {
  name: string;
  onName: (name: string) => void;
  types: FilterOption[];
  selectedTypes: string[];
  onToggleType: (type: string) => void;
};

// The left panel of favorites.png: "Product name" search, then "Type" with
// counts. It filters the list already on screen, so there is no request.
// Reuses the store's filter pieces so both panels look the same.
export function WishlistFilters({ name, onName, types, selectedTypes, onToggleType }: Props) {
  return (
    <aside aria-label="Filters" className="self-start bg-brand-mid">
      <section className="border-b-2 border-brand px-5 py-5 last:border-b-0">
        <label htmlFor="wishlist-name" className="text-base font-bold">
          Product name
        </label>
        <div className="relative mt-4">
          <input
            id="wishlist-name"
            type="text"
            value={name}
            onChange={(e) => onName(e.target.value)}
            className="h-11 w-full border border-brand-border bg-brand pr-10 pl-3 text-base text-white"
          />
          <Search className="pointer-events-none absolute top-3 right-3 size-5" aria-hidden />
        </div>
      </section>

      {/* Hidden while the list is empty, as on Eneba. */}
      {types.length > 0 && (
        <FilterSection title="Type" active={selectedTypes.length}>
          <CheckList options={types} selected={selectedTypes} onToggle={onToggleType} />
        </FilterSection>
      )}
    </aside>
  );
}
