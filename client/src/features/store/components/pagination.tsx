"use client";

import { ChevronLeft, ChevronRight } from "lucide-react";
import { useStoreParams } from "../hooks/use-store-params";

type Props = { page: number; pageSize: number; total: number };

const BOX = "flex size-10 items-center justify-center text-base font-bold";
const IDLE = "border border-white/30 transition-colors hover:bg-brand-darkest disabled:opacity-40";

export function Pagination({ page, pageSize, total }: Props) {
  const { update } = useStoreParams();
  const pages = Math.ceil(total / pageSize);
  if (pages <= 1) return null;

  // Page 1 has no param, so its URL matches the plain store link.
  // scroll: a new page starts at the top, unlike a filter change.
  const go = (p: number) => update({ page: p === 1 ? null : String(p) }, true);

  return (
    <nav aria-label="Pages" className="mt-10 flex flex-wrap items-center justify-center gap-2">
      <button
        type="button"
        disabled={page <= 1}
        onClick={() => go(page - 1)}
        aria-label="Previous page"
        className={`${BOX} ${IDLE}`}
      >
        <ChevronLeft className="size-5" aria-hidden />
      </button>

      {pageList(page, pages).map((p, i) =>
        p === null ? (
          <span key={`gap-${i}`} className="px-1 text-text-muted" aria-hidden>
            …
          </span>
        ) : (
          <button
            key={p}
            type="button"
            onClick={() => go(p)}
            aria-current={p === page ? "page" : undefined}
            className={`${BOX} ${p === page ? "bg-yellow text-ink" : IDLE}`}
          >
            {p}
          </button>
        ),
      )}

      <button
        type="button"
        disabled={page >= pages}
        onClick={() => go(page + 1)}
        aria-label="Next page"
        className={`${BOX} ${IDLE}`}
      >
        <ChevronRight className="size-5" aria-hidden />
      </button>
    </nav>
  );
}

// First, last, and two either side of the current page; null is a "…" gap.
// Page 6 of 12 → 1 … 4 5 6 7 8 … 12
function pageList(page: number, pages: number): (number | null)[] {
  const out: (number | null)[] = [];
  let last = 0;
  for (let p = 1; p <= pages; p++) {
    if (p !== 1 && p !== pages && Math.abs(p - page) > 2) continue;
    if (p - last > 1) out.push(null);
    out.push(p);
    last = p;
  }
  return out;
}
