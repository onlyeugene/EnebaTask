"use client";

import { useState } from "react";
import { Check, ChevronDown, ChevronUp } from "lucide-react";
import type { FilterOption } from "@/types/api";

type Props = {
  options: FilterOption[];
  selected: string[];
  onToggle: (value: string) => void;
  // Show this many, then an "N more" link (Eneba's "3 more").
  limit?: number;
};

export function CheckList({ options, selected, onToggle, limit }: Props) {
  const [showAll, setShowAll] = useState(false);
  // A ticked option always shows, even past the limit, so it can be unticked.
  const shown =
    showAll || !limit
      ? options
      : options.filter((o, i) => i < limit || selected.includes(o.value));
  const hidden = options.length - shown.length;

  return (
    <>
      <ul className="flex flex-col gap-3">
        {shown.map((o) => {
          const checked = selected.includes(o.value);
          return (
            <li key={o.value}>
              <label className="flex cursor-pointer items-center gap-3 text-base">
                {/* The real checkbox is hidden but still focusable and read
                    out; the square next to it is the visible box. */}
                <input
                  type="checkbox"
                  checked={checked}
                  onChange={() => onToggle(o.value)}
                  className="peer sr-only"
                />
                <span
                  aria-hidden
                  className="flex size-5 shrink-0 items-center justify-center border-2 border-white peer-checked:border-yellow peer-checked:bg-yellow peer-focus-visible:outline-2 peer-focus-visible:outline-offset-2 peer-focus-visible:outline-white"
                >
                  {checked && <Check className="size-4 text-ink" strokeWidth={3} />}
                </span>
                <span className="flex-1">{o.label}</span>
                <span className="text-text-secondary">{o.count}</span>
              </label>
            </li>
          );
        })}
      </ul>

      {hidden > 0 && (
        <button
          type="button"
          onClick={() => setShowAll(true)}
          className="mt-3 flex items-center gap-1 text-base font-bold text-yellow"
        >
          {hidden} more <ChevronDown className="size-4" aria-hidden />
        </button>
      )}
      {showAll && limit !== undefined && options.length > limit && (
        <button
          type="button"
          onClick={() => setShowAll(false)}
          className="mt-3 flex items-center gap-1 text-base font-bold text-yellow"
        >
          Show less <ChevronUp className="size-4" aria-hidden />
        </button>
      )}
    </>
  );
}
