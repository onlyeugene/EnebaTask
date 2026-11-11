"use client";

import type { ComponentType } from "react";
import { useSearchParams } from "next/navigation";
import type { BlockType, LayoutEntry } from "@/types/api";
import { ErrorBoundary } from "@/components/ui/error-boundary";
import { ErrorState } from "@/components/ui/error-state";
import { CardSkeleton, Skeleton } from "@/components/ui/skeleton";
import { useHomepage } from "../hooks/use-homepage";
import { BlockError } from "./block-shell";
import { CategoryGridBlock } from "./category-grid-block";
import { CategoryRowBlock } from "./category-row-block";
import { HeroBlock } from "./hero-block";
import { ProductRailBlock } from "./product-rail-block";

type BlockProps = { entry: LayoutEntry; fail: boolean };

// Block type → component. A new block type needs one entry here.
const COMPONENTS: Record<BlockType, ComponentType<BlockProps>> = {
  hero: HeroBlock,
  category_row: CategoryRowBlock,
  category_grid: CategoryGridBlock,
  bestsellers: ProductRailBlock,
  deals: ProductRailBlock,
  new_releases: ProductRailBlock,
  gift_cards: ProductRailBlock,
};

// Full-width background bands, alternating lighter and darkest purple.
// Each band also sets --band to its color, for things that must blend into
// it (the category row's edge fade). py-10 is the 40px section spacing from
// notes.md, kept inside the band so the colors touch.
const BANDS = [
  "bg-brand-mid [--band:var(--color-brand-mid)]",
  "bg-[#3c1395] [--band:#3c1395]",
];
const BAND = "py-10";

// The page order and block list come from GET /api/homepage; each block
// then loads its own content. Nothing about the content is fixed here.
export function HomeBlocks() {
  const layout = useHomepage();
  // Demo switch: /?fail=<blockId> makes that one block's request fail.
  const failId = useSearchParams().get("fail");

  if (layout.isPending || (layout.isError && layout.isFetching)) return <HomeSkeleton />;

  if (layout.isError) {
    return (
      <div className="mx-auto max-w-page px-4 py-10">
        <ErrorState
          message="The homepage couldn't load."
          onRetry={() => {
            void layout.refetch();
          }}
        />
      </div>
    );
  }

  // A type this client doesn't know yet (newer API) is skipped, not a crash,
  // and is left out before counting bands so the colors still alternate.
  const blocks = layout.data.blocks.filter((entry) => COMPONENTS[entry.type]);
  // The hero runs edge to edge in its own colors, so bands count the other
  // blocks only, in layout order. Adding, removing or reordering blocks in
  // the db keeps the colors alternating.
  let band = 0;

  return blocks.map((entry, i) => {
    const Block = COMPONENTS[entry.type];
    const content = (
      // The boundary catches render bugs; BlockShell handles request errors.
      <ErrorBoundary fallback={(reset) => <BlockError title={entry.title} onRetry={reset} />}>
        <Block entry={entry} fail={entry.id === failId} />
      </ErrorBoundary>
    );
    if (entry.type === "hero") return <div key={entry.id}>{content}</div>;

    // The category row has no heading; it leads into the block after it, so
    // both share one band. It keeps only its top padding, and the next
    // block's top padding gives the usual 40px between them.
    const leadsIn = entry.type === "category_row" && i < blocks.length - 1;
    const bg = BANDS[band % BANDS.length];
    if (!leadsIn) band++;
    return (
      <div key={entry.id} className={`${bg} ${leadsIn ? "pt-10" : BAND}`}>
        {content}
      </div>
    );
  });
}

// Shown before the layout arrives: a hero-sized box and one rail in the first band.
export function HomeSkeleton() {
  return (
    <div aria-busy>
      <Skeleton className="h-115 bg-brand-da md:h-90" />
      <div className={`${BANDS[0]} ${BAND}`}>
        <div className="mx-auto max-w-page px-4">
          <Skeleton className="mb-6 h-9.5 w-75 max-w-full bg-brand-light/40" />
          <div className="flex gap-5 overflow-hidden">
            {Array.from({ length: 5 }, (_, i) => (
              <CardSkeleton key={i} />
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
