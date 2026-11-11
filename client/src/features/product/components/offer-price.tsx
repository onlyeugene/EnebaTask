import { CirclePlus } from "lucide-react";
import { discountLabel, discountPct } from "@/lib/format";
import { Price } from "@/components/ui/price";

type Props = {
  priceCents: number;
  originalCents: number;
  cashbackPct: number;
  align?: "left" | "right";
};

// "€24.99 -49%", the price, then "14% Cashback", as on each Eneba offer.
// Cashback is a product setting (cashbackPct), the same for every seller.
export function OfferPrice({ priceCents, originalCents, cashbackPct, align = "right" }: Props) {
  const pct = discountPct(originalCents, priceCents);
  const end = align === "right" ? "items-end text-right" : "items-start";

  return (
    <div className={`flex flex-col gap-1 ${end}`}>
      {pct > 0 && (
        <p className="text-sm font-bold">
          <Price cents={originalCents} className="text-text-muted line-through" />{" "}
          <span className="text-lime">{discountLabel(pct)}</span>
        </p>
      )}
      <Price cents={priceCents} className="text-3xl font-extrabold" />
      {cashbackPct > 0 && (
        <p className="flex items-center gap-1.5 text-sm font-bold text-lime">
          <CirclePlus className="size-4" aria-hidden />
          {cashbackPct}% Cashback
        </p>
      )}
    </div>
  );
}
