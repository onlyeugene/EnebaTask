import type { Platform, ProductType, Region } from "@/types/api";

const TYPE_LABELS: Record<ProductType, string> = {
  game: "Game",
  gift_card: "Gift card",
  software: "Software",
};

export function typeLabel(type: ProductType): string {
  return TYPE_LABELS[type];
}

// 1549 → "€15.49". Integer math only, so no float rounding can creep in.
export function formatCents(cents: number): string {
  const euros = Math.floor(cents / 100);
  const rest = String(cents % 100).padStart(2, "0");
  return `€${euros}.${rest}`;
}

// 1250 → "12.50", for price input boxes.
export function centsToEuros(cents: number): string {
  return formatCents(cents).slice(1);
}

// "12", "12.5" or "12,50" → cents. Built from the digits with integer math,
// never parseFloat. null when the text isn't a price.
export function eurosToCents(text: string): number | null {
  const m = /^(\d{1,7})(?:[.,](\d{1,2}))?$/.exec(text.trim());
  if (!m) return null;
  return Number(m[1]) * 100 + Number((m[2] ?? "0").padEnd(2, "0"));
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
