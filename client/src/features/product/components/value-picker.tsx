"use client";

import Link from "next/link";
import { CircleCheck } from "lucide-react";
import type { ProductDetail } from "@/types/api";
import { discountLabel } from "@/lib/format";
import { Price } from "@/components/ui/price";
import { useGiftCardFamily } from "../hooks/use-gift-card-family";

// Gift cards only: one tile per value of the same card (Eneba's "Value"
// grid). Each value is its own product, so a tile is a link to it. Hidden
// while loading, on error, and when the card comes in only one value: the
// offers below still work without it.
export function ValuePicker({ product }: { product: ProductDetail }) {
  const family = useGiftCardFamily(product);
  if (product.type !== "gift_card" || !family.data || family.data.length < 2) return null;

  return (
    <section>
      <p className="mb-4 text-base">
        <b>Value:</b> <span className="text-text-secondary">{product.giftCardValue}</span>
      </p>
      <ul className="grid grid-cols-[repeat(auto-fill,minmax(140px,1fr))] gap-3">
        {family.data.map((p) => {
          const current = p.id === product.id;
          return (
            <li key={p.id}>
              <Link
                href={`/product/${p.slug}`}
                aria-current={current ? "page" : undefined}
                className={`relative flex h-full flex-col border bg-brand-dark transition-colors hover:border-white ${current ? "border-white" : "border-transparent"}`}
              >
                <div className="bg-brand-darkest/40 px-3 py-3">
                  <p className="text-xl font-extrabold">{p.giftCardValue}</p>
                  {p.cashbackPct > 0 && (
                    <p className="mt-1 text-xs font-bold text-lime">{p.cashbackPct}% Cashback</p>
                  )}
                </div>
                <div className="px-3 py-3 text-sm">
                  {p.minPriceCents === null ? (
                    <p className="text-text-muted">Sold out</p>
                  ) : (
                    <>
                      {p.discountPct > 0 && (
                        <p className="text-xs font-bold">
                          <Price cents={p.originalPriceCents} className="text-text-muted line-through" />{" "}
                          <span className="text-lime">{discountLabel(p.discountPct)}</span>
                        </p>
                      )}
                      <Price cents={p.minPriceCents} className="font-bold" />
                    </>
                  )}
                </div>
                {current && (
                  <CircleCheck
                    className="absolute right-1 bottom-1 size-4 fill-teal text-white"
                    aria-hidden
                  />
                )}
              </Link>
            </li>
          );
        })}
      </ul>
    </section>
  );
}
