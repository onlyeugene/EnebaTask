"use client";

import { Star } from "lucide-react";
import type { Offer } from "@/types/api";
import { useOfferActions } from "../hooks/use-offer-actions";
import { AddButton } from "./add-button";
import { OfferPrice } from "./offer-price";

type Props = { offer: Offer; originalCents: number; cashbackPct: number };

// One seller's offer (product_details2.png). As on Eneba, the "+cart" button
// slides in and "Buy now" turns yellow while the row is hovered. Touch
// screens can't hover, so they always get both buttons.
export function OfferRow({ offer, originalCents, cashbackPct }: Props) {
  const { add, buyNow } = useOfferActions();
  const soldOut = offer.stock === 0;

  return (
    <li className="group flex flex-col gap-4 bg-brand-dark p-5 sm:flex-row sm:items-center sm:justify-between">
      <div className="flex items-center gap-4">
        {/* A letter tile instead of a seller logo. */}
        <span className="flex size-12 shrink-0 items-center justify-center bg-brand-light text-xl font-extrabold">
          {offer.seller.name.charAt(0)}
        </span>
        <div>
          <p className="text-base font-semibold">{offer.seller.name}</p>
          <p className="mt-1 flex items-center gap-1.5 text-sm text-text-secondary">
            <Star className="size-4 fill-yellow text-yellow" aria-hidden />
            <b className="text-white">{offer.seller.rating.toFixed(1)}</b> rating ·{" "}
            {offer.seller.salesCount.toLocaleString("en-US")} sales
          </p>
        </div>
      </div>

      <div className="flex items-center justify-between gap-5 sm:justify-end">
        {soldOut ? (
          <p className="text-xl font-extrabold text-text-muted">Sold out</p>
        ) : (
          <>
            <OfferPrice
              priceCents={offer.priceCents}
              originalCents={originalCents}
              cashbackPct={cashbackPct}
            />
            <div className="flex items-center">
              {/* Touch screens: always shown. With a mouse: 0 wide until the row
                  is hovered (or focused by keyboard), then it opens from the
                  right edge, so the button seems to slide out from behind "Buy now". */}
              <div className="flex justify-end overflow-hidden transition-[max-width,opacity] duration-200 can-hover:max-w-0 can-hover:opacity-0 can-hover:group-hover:max-w-20 can-hover:group-hover:opacity-100 can-hover:group-focus-within:max-w-20 can-hover:group-focus-within:opacity-100">
                <AddButton onClick={() => add(offer.id)} className="mr-3" />
              </div>
              <button
                type="button"
                onClick={() => buyNow(offer.id)}
                // Outline at rest, solid yellow with dark text while the row is
                // hovered or focused. No hover style of its own: the shared
                // outline style's hover background used to win over the row's
                // yellow and left dark text on a dark box.
                className="inline-flex h-12 shrink-0 items-center justify-center border border-white px-6 text-sm font-bold text-white transition-colors group-hover:border-yellow group-hover:bg-yellow group-hover:text-ink group-focus-within:border-yellow group-focus-within:bg-yellow group-focus-within:text-ink"
              >
                Buy now
              </button>
            </div>
          </>
        )}
      </div>
    </li>
  );
}
