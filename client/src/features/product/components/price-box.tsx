"use client";

import type { ReactNode } from "react";
import { CirclePlus, LockKeyhole } from "lucide-react";
import type { ProductDetail } from "@/types/api";
import { discountLabel } from "@/lib/format";
import { Button } from "@/components/ui/button";
import { Price } from "@/components/ui/price";

// The right-hand box (Eneba's product page): the lowest price with "From",
// and "See offers", which scrolls to the offer list where the buying happens.
// The product's minPriceCents and discountPct are rebuilt from live stock by
// the API, so they match the cheapest offer below. Eneba's PayPal button,
// payment logos and trust badges are left out (brand logos); the second box
// says plainly that checkout is simulated.
export function PriceBox({ product }: { product: ProductDetail }) {
  const price = product.minPriceCents;
  const seeOffers = () =>
    document.getElementById("offers")?.scrollIntoView({ behavior: "smooth" });

  return (
    <aside className="flex flex-col gap-2">
      <div className="bg-brand-mid p-6">
        {price === null ? (
          <>
            <p className="text-3xl font-extrabold text-text-muted">Sold out</p>
            <p className="mt-2 text-sm text-text-secondary">
              No seller has keys for this product right now.
            </p>
          </>
        ) : (
          <>
            <p className="text-sm font-bold text-text-muted">
              From:{" "}
              {product.discountPct > 0 && (
                <>
                  <Price cents={product.originalPriceCents} className="line-through" />{" "}
                  <span className="text-lime">{discountLabel(product.discountPct)}</span>
                </>
              )}
            </p>
            <Price cents={price} className="mt-1 block text-3xl font-extrabold" />
            {product.cashbackPct > 0 && (
              <p className="mt-2 flex items-center gap-1.5 text-sm font-bold text-lime">
                <CirclePlus className="size-4" aria-hidden />
                {product.cashbackPct}% Cashback
              </p>
            )}
            <Button className="mt-6 h-12 w-full text-base" onClick={seeOffers}>
              See offers
            </Button>
          </>
        )}
      </div>

      <div className="bg-brand-mid p-6 text-sm">
        <Point icon={<LockKeyhole className="size-5" />}>
          Simulated checkout, no real payments or personal data are processed.
        </Point>
      </div>
    </aside>
  );
}

function Point({ icon, children }: { icon: ReactNode; children: ReactNode }) {
  return (
    <p className="flex items-center gap-3 text-text-secondary">
      <span className="shrink-0 text-mint" aria-hidden>
        {icon}
      </span>
      {children}
    </p>
  );
}
