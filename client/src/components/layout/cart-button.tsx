"use client";

import { ShoppingCart } from "lucide-react";

type Props = { count: number; open: boolean; onClick: () => void };

// Cart icon with Eneba's red count badge. Display only: the count and what a
// click does come from the cart feature (CartPopover), so layout code never
// imports a feature.
export function CartButton({ count, open, onClick }: Props) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-expanded={open}
      aria-label={`Cart, ${count} ${count === 1 ? "item" : "items"}`}
      className={`link-hover relative p-2.5 ${open ? "text-yellow" : ""}`}
    >
      <ShoppingCart className="size-7" strokeWidth={1.75} />
      {count > 0 && (
        // The radius tokens are reset (square corners), so the circle is set by hand.
        <span className="absolute top-1 right-0.5 flex h-4.5 min-w-4.5 items-center justify-center rounded-[50%] bg-red px-1 text-2xs font-bold text-white">
          {count > 99 ? "99+" : count}
        </span>
      )}
    </button>
  );
}
