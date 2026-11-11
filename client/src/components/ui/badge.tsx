import { CircleDollarSign } from "lucide-react";

// Mint "CASHBACK" pill that sits on the bottom left of a product cover.
export function CashbackBadge() {
  return (
    <span className="inline-flex items-center gap-1 bg-mint px-2 py-1 text-xs font-bold uppercase text-ink">
      <CircleDollarSign className="size-3.5" aria-hidden />
      Cashback
    </span>
  );
}

// "GIFT CARD | €50" tag for gift cards, top right of the cover.
export function GiftCardTag({ value }: { value: string }) {
  return (
    <span className="flex text-2xs font-extrabold uppercase">
      <span className="bg-brand-darkest px-1.5 py-1 text-white">Gift card</span>
      <span className="bg-yellow px-1.5 py-1 text-ink">{value}</span>
    </span>
  );
}
