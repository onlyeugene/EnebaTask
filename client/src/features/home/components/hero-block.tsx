"use client";

import Link from "next/link";
import { ArrowRight, KeyRound } from "lucide-react";
import type { Banner, HeroData, LayoutEntry } from "@/types/api";
import { buttonStyles } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { BlockShell } from "./block-shell";

// The API sends a theme name; the colors stay on the client, from the tokens.
const THEMES: Record<Banner["theme"], string> = {
  purple: "bg-linear-to-br from-brand-light to-brand-dark text-white",
  yellow: "bg-yellow text-ink",
  mint: "bg-mint text-ink",
  pink: "bg-pink text-white",
};

// Main banner full width on phones with the tiles in a row below;
// from md up, Eneba's layout: wide banner left, three narrow tiles right.
const GRID = "grid grid-cols-3 md:grid-cols-[minmax(0,1fr)_repeat(3,180px)] lg:grid-cols-[minmax(0,1fr)_repeat(3,220px)]";

export function HeroBlock({ entry, fail }: { entry: LayoutEntry; fail: boolean }) {
  return (
    <BlockShell<HeroData> entry={entry} fail={fail} skeleton={<HeroSkeleton />}>
      {({ data }) => (
        <section aria-label="Featured" className={GRID}>
          <MainBanner banner={data.main} />
          {data.tiles.map((tile) => (
            <Tile key={tile.href} banner={tile} />
          ))}
        </section>
      )}
    </BlockShell>
  );
}

function MainBanner({ banner }: { banner: Banner }) {
  return (
    <Link
      href={banner.href}
      className={`relative col-span-3 flex h-[260px] items-center overflow-hidden md:col-span-1 md:h-[360px] ${THEMES[banner.theme]}`}
    >
      {/* No artwork (no copyrighted images), so a large faint icon fills the space. */}
      <KeyRound
        aria-hidden
        strokeWidth={1}
        className="absolute -right-10 -bottom-16 size-[320px] rotate-12 opacity-15 md:size-[440px]"
      />
      <div className="relative mx-auto w-full max-w-[760px] px-6">
        <h2 className="text-3xl font-extrabold uppercase md:text-[56px] md:leading-[60px]">
          {banner.title}
        </h2>
        <p className="mt-3 max-w-md text-base font-semibold opacity-90">{banner.subtitle}</p>
        <span className={`${buttonStyles.primary} mt-6 min-w-48 text-base uppercase`}>
          {banner.ctaLabel}
        </span>
      </div>
    </Link>
  );
}

function Tile({ banner }: { banner: Banner }) {
  return (
    <Link
      href={banner.href}
      className={`flex h-[200px] flex-col justify-between p-4 transition-[filter] hover:brightness-110 md:h-[360px] md:p-5 ${THEMES[banner.theme]}`}
    >
      <div>
        <h2 className="text-xl font-extrabold uppercase md:text-3xl">{banner.title}</h2>
        <p className="mt-2 text-xs font-semibold opacity-90 md:text-sm">{banner.subtitle}</p>
      </div>
      <span className="flex items-center gap-1 text-sm font-bold">
        {banner.ctaLabel}
        <ArrowRight className="size-4" aria-hidden />
      </span>
    </Link>
  );
}

function HeroSkeleton() {
  return (
    <div className={GRID}>
      <Skeleton className="col-span-3 h-[260px] bg-brand-dark md:col-span-1 md:h-[360px]" />
      <Skeleton className="h-[200px] bg-brand-darkest md:h-[360px]" />
      <Skeleton className="h-[200px] bg-brand-dark md:h-[360px]" />
      <Skeleton className="h-[200px] bg-brand-darkest md:h-[360px]" />
    </div>
  );
}
