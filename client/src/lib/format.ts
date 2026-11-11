import type { Platform, Region } from "@/types/api";

// 1549 → "€15.49". Integer math only, so no float rounding can creep in.
export function formatCents(cents: number): string {
  const euros = Math.floor(cents / 100);
  const rest = String(cents % 100).padStart(2, "0");
  return `€${euros}.${rest}`;
}

// Same rule as the server's cart lines: cashback is rounded down.
export function cashbackCents(priceCents: number, pct: number): number {
  return Math.floor((priceCents * pct) / 100);
}

export function discountLabel(pct: number): string {
  return `-${pct}%`;
}

const REGION_LABELS: Record<Region, string> = {
  global: "GLOBAL",
  europe: "EUROPE",
  united_states: "UNITED STATES",
  united_kingdom: "UNITED KINGDOM",
};

export function regionLabel(region: Region): string {
  return REGION_LABELS[region];
}

// Eneba colors GLOBAL and EUROPE teal, country regions pink.
export function regionColor(region: Region): string {
  return region === "global" || region === "europe" ? "text-teal" : "text-pink";
}

const PLATFORM_LABELS: Record<Platform, string> = {
  steam: "Steam",
  psn: "PSN",
  xbox: "Xbox",
  nintendo: "Nintendo",
  windows: "Windows",
};

export function platformLabel(platform: Platform): string {
  return PLATFORM_LABELS[platform];
}
