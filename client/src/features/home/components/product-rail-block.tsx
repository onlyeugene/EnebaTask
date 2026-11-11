"use client";

import Link from "next/link";
import { Info } from "lucide-react";
import type { LayoutEntry, RailData } from "@/types/api";
import { buttonStyles } from "@/components/ui/button";
import { ProductCard } from "@/components/ui/product-card";
import { CardSkeleton, Skeleton } from "@/components/ui/skeleton";
import { BlockShell } from "./block-shell";
import { ScrollRow } from "./scroll-row";

// One component for all four rails (bestsellers, deals, new releases, gift
// cards): the server decides which products and which "See all" link.
export function ProductRailBlock({ entry, fail }: { entry: LayoutEntry; fail: boolean }) {
  return (
    <BlockShell<RailData> entry={entry} fail={fail} skeleton={<RailSkeleton />}>
      {({ data, title }) => (
        // group/rail: the rail's arrows show while the mouse is anywhere on this section.
        <section className="group/rail mx-auto max-w-page px-4">
          <h2 className="mb-6 flex items-center gap-2 text-3xl font-extrabold">
            {title}
            <Info className="size-5 text-text-muted" aria-hidden />
          </h2>

          {data.products.length === 0 ? (
            <p className="text-base text-text-muted">Nothing here right now. Check back soon.</p>
          ) : (
            <ScrollRow label={title} variant="rail">
              {data.products.map((p) => (
                <ProductCard key={p.id} product={p} />
              ))}
            </ScrollRow>
          )}

          <div className="mt-8 flex justify-center">
            <Link href={data.seeAllHref} className={`${buttonStyles.outline} w-64 text-base`}>
              See all
            </Link>
          </div>
        </section>
      )}
    </BlockShell>
  );
}

// Light bar for the title and dark cards with a spinner, as on Eneba.
function RailSkeleton() {
  return (
    <div className="mx-auto max-w-page px-4">
      <Skeleton className="mb-6 h-[38px] w-[300px] max-w-full bg-brand-light/40" />
      <div className="flex gap-5 overflow-hidden">
        {Array.from({ length: 5 }, (_, i) => (
          <CardSkeleton key={i} />
        ))}
      </div>
    </div>
  );
}
