"use client";

import { useState } from "react";
import Link from "next/link";
import type { CategoryRowData, LayoutEntry } from "@/types/api";
import { buttonStyles } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { BlockShell } from "./block-shell";
import { CategoryIcon } from "./category-icon";

// One row on desktop; "Load more" shows the next row. The API already sent
// every tile, so this only changes how many are drawn.
const PAGE = 7;

const GRID = "grid grid-cols-2 gap-5 xs:grid-cols-3 sm:grid-cols-4 md:grid-cols-5 lg:grid-cols-7";

export function CategoryGridBlock({ entry, fail }: { entry: LayoutEntry; fail: boolean }) {
  const [shown, setShown] = useState(PAGE);

  return (
    <BlockShell<CategoryRowData> entry={entry} fail={fail} skeleton={<GridSkeleton />}>
      {({ data, title }) => (
        <section className="mx-auto max-w-page px-4">
          <h2 className="mb-6 text-3xl font-extrabold">{title}</h2>
          {/* Same line as an empty rail (no product in stock at all). */}
          {data.categories.length === 0 && (
            <p className="text-base text-text-muted">Nothing here right now. Check back soon.</p>
          )}
          <div className={GRID}>
            {data.categories.slice(0, shown).map((c) => (
              // White at low opacity stays visible on either band color.
              <Link
                key={c.slug}
                href={c.href}
                className="flex h-[140px] flex-col items-center justify-between bg-white/10 py-5 transition-colors hover:bg-brand-dark"
              >
                <span className="text-sm font-bold">{c.count}</span>
                <CategoryIcon name={c.icon} className="size-11" />
                <span className="text-base font-bold">{c.label}</span>
              </Link>
            ))}
          </div>
          {shown < data.categories.length && (
            <div className="mt-8 flex justify-center">
              <button
                type="button"
                onClick={() => setShown((n) => n + PAGE)}
                className={`${buttonStyles.outline} w-64 text-base`}
              >
                Load more
              </button>
            </div>
          )}
        </section>
      )}
    </BlockShell>
  );
}

function GridSkeleton() {
  return (
    <div className="mx-auto max-w-page px-4">
      <Skeleton className="mb-6 h-[38px] w-[240px] max-w-full bg-brand-light/40" />
      <div className={GRID}>
        {Array.from({ length: PAGE }, (_, i) => (
          <Skeleton key={i} className="h-[140px] bg-white/10" />
        ))}
      </div>
    </div>
  );
}
