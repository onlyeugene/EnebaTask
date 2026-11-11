"use client";

import { useState, type FormEvent } from "react";
import { centsToEuros, eurosToCents } from "@/lib/format";

type Props = {
  minCents: string | null; // from the URL
  maxCents: string | null;
  bounds: { minCents: number; maxCents: number };
  onApply: (minCents: string | null, maxCents: string | null) => void;
};

// Two euro boxes. The price applies on Enter, or when focus leaves both
// boxes, not on every keystroke (typing "15" would first search for "1").
export function PriceRange({ minCents, maxCents, bounds, onApply }: Props) {
  const [min, setMin] = useState(minCents ? centsToEuros(Number(minCents)) : "");
  const [max, setMax] = useState(maxCents ? centsToEuros(Number(maxCents)) : "");
  const [error, setError] = useState("");

  function apply() {
    const lo = min.trim() ? eurosToCents(min) : null;
    const hi = max.trim() ? eurosToCents(max) : null;
    if ((min.trim() && lo === null) || (max.trim() && hi === null)) {
      setError("Enter a price like 12.50");
      return;
    }
    // Checked here too, so the API never gets a pair it answers with a 400.
    if (lo !== null && hi !== null && lo > hi) {
      setError("Min can't be above max");
      return;
    }
    setError("");
    const nextMin = lo === null ? null : String(lo);
    const nextMax = hi === null ? null : String(hi);
    if (nextMin !== minCents || nextMax !== maxCents) onApply(nextMin, nextMax);
  }

  function onSubmit(e: FormEvent) {
    e.preventDefault();
    apply();
  }

  const box =
    "h-11 w-full min-w-0 border border-brand-border bg-brand px-2 text-base placeholder:text-text-faint";

  return (
    <form
      onSubmit={onSubmit}
      // Moving from the min box to the max box is not "leaving".
      onBlur={(e) => {
        if (!e.currentTarget.contains(e.relatedTarget)) apply();
      }}
    >
      <div className="flex items-center gap-2">
        <input
          aria-label="Minimum price in euros"
          inputMode="decimal"
          value={min}
          onChange={(e) => setMin(e.target.value)}
          placeholder={centsToEuros(bounds.minCents)}
          className={box}
        />
        <span aria-hidden>-</span>
        <input
          aria-label="Maximum price in euros"
          inputMode="decimal"
          value={max}
          onChange={(e) => setMax(e.target.value)}
          placeholder={centsToEuros(bounds.maxCents)}
          className={box}
        />
      </div>
      {error && (
        <p role="alert" className="mt-2 text-xs text-red-pink">
          {error}
        </p>
      )}
      {/* A form with two inputs only submits on Enter if it has a submit button. */}
      <button type="submit" className="sr-only">
        Apply price
      </button>
    </form>
  );
}
