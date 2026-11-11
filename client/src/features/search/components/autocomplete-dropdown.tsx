"use client";

import { useEffect } from "react";
import type { UseQueryResult } from "@tanstack/react-query";
import type { Suggestion } from "@/types/api";
import { API_URL } from "@/lib/api-url";
import { typeLabel } from "@/lib/format";
import { buttonStyles } from "@/components/ui/button";
import { Price } from "@/components/ui/price";

type Props = {
  id: string;
  text: string;
  query: UseQueryResult<Suggestion[]>;
  active: number;
  onHover: (index: number) => void;
  onPick: (href: string) => void;
};

const NOTE = "px-5 py-4 text-base text-text-secondary";

// Eneba's search dropdown: product rows, then "Show all results".
export function AutocompleteDropdown({ id, text, query, active, onHover, onPick }: Props) {
  const items = query.data ?? [];

  // Arrow keys can move past the visible part of a scrolled list.
  useEffect(() => {
    if (active >= 0) document.getElementById(`${id}-${active}`)?.scrollIntoView({ block: "nearest" });
  }, [id, active]);

  return (
    // preventDefault on mousedown keeps focus in the text box, so its blur
    // doesn't close this list before the click lands.
    // Placed 1px below the box so its bottom border stays visible. On phones
    // it runs edge to edge: the box sits 16px in from the screen, plus its border.
    <div
      onMouseDown={(e) => e.preventDefault()}
      className="absolute -inset-x-[17px] top-[calc(100%+1px)] z-50 max-h-[75vh] overflow-y-auto bg-brand md:-inset-x-px"
    >
      {query.isError ? (
        <p className={NOTE}>Suggestions are unavailable right now. Press Enter to search.</p>
      ) : items.length === 0 ? (
        // Pending also covers the debounce before the first request.
        <p className={NOTE}>
          {query.isPending || query.isFetching ? "Searching…" : `No results for “${text}”`}
        </p>
      ) : (
        <ul id={id} role="listbox" aria-label="Suggestions">
          {items.map((s, i) => (
            <Row
              key={s.slug}
              id={`${id}-${i}`}
              suggestion={s}
              active={i === active}
              onHover={() => onHover(i)}
              onPick={() => onPick(`/product/${s.slug}`)}
            />
          ))}
        </ul>
      )}

      {/* Hidden with no rows: "Show all" would open the same empty search. */}
      {(items.length > 0 || query.isError) && (
        <div className="p-5">
          <button
            type="button"
            onClick={() => onPick(`/store?q=${encodeURIComponent(text)}`)}
            className={`${buttonStyles.outline} w-full text-base`}
          >
            Show all results
          </button>
        </div>
      )}
    </div>
  );
}

type RowProps = {
  id: string;
  suggestion: Suggestion;
  active: boolean;
  onHover: () => void;
  onPick: () => void;
};

// Mouse and arrow keys share one highlight (`active`), so the two never
// light up different rows.
function Row({ id, suggestion: s, active, onHover, onPick }: RowProps) {
  const price = s.minPriceCents;
  return (
    <li
      id={id}
      role="option"
      aria-selected={active}
      onMouseEnter={onHover}
      onClick={onPick}
      className={`flex cursor-pointer items-center gap-5 border-t border-divider px-5 py-3 first:border-t-0 ${active ? "bg-brand-mid" : ""}`}
    >
      {/* eslint-disable-next-line @next/next/no-img-element -- SVG cover, see product-card.tsx */}
      <img
        src={API_URL + s.coverUrl}
        alt=""
        width={60}
        height={84}
        className="h-[84px] w-[60px] shrink-0 object-cover"
      />

      <div className="flex min-w-0 flex-1 flex-col items-start gap-2">
        <span className="rounded-full border border-text-muted px-2.5 text-xs leading-5 text-text-secondary">
          {typeLabel(s.type)}
        </span>
        <span className="line-clamp-2 text-base font-bold">{s.name}</span>
      </div>

      <div className="flex shrink-0 flex-col items-end text-right">
        {price === null ? (
          <span className="text-base font-extrabold text-text-muted">Sold out</span>
        ) : (
          <>
            <span className="text-sm font-semibold text-text-muted">
              From{" "}
              {s.discountPct > 0 && (
                <Price cents={s.originalPriceCents} className="line-through" />
              )}
            </span>
            <Price cents={price} className="text-xl font-extrabold" />
          </>
        )}
      </div>
    </li>
  );
}
