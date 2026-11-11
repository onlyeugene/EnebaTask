"use client";

import Link from "next/link";
import type { CategoryRowData, LayoutEntry } from "@/types/api";
import { Skeleton } from "@/components/ui/skeleton";
import { BlockShell } from "./block-shell";
import { CategoryIcon } from "./category-icon";
import { ScrollRow } from "./scroll-row";

const ITEM = "flex w-[105px] shrink-0 flex-col items-center gap-2 py-3";

export function CategoryRowBlock({ entry, fail }: { entry: LayoutEntry; fail: boolean }) {
  return (
    <BlockShell<CategoryRowData> entry={entry} fail={fail} skeleton={<CategorySkeleton />}>
      {({ data, title }) => (
        // No visible heading, like Eneba; the title is kept for screen readers.
        <section aria-label={title} className="mx-auto max-w-page px-4">
          <ScrollRow label={title} variant="categories">
            {data.categories.map((c) => (
              <Link key={c.slug} href={c.href} className={`${ITEM} link-hover`}>
                <CategoryIcon name={c.icon} className="size-9" />
                <span className="text-sm font-bold">{c.label}</span>
                <span className="text-xs text-text-faint">{c.count} products</span>
              </Link>
            ))}
          </ScrollRow>
        </section>
      )}
    </BlockShell>
  );
}

function CategorySkeleton() {
  return (
    <div className="mx-auto flex max-w-page gap-5 overflow-hidden px-4">
      {Array.from({ length: 8 }, (_, i) => (
        <div key={i} className={ITEM}>
          <Skeleton className="size-9 bg-brand-light/40" />
          <Skeleton className="h-4 w-16 bg-brand-light/40" />
        </div>
      ))}
    </div>
  );
}
