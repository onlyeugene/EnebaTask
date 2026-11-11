"use client";

import Link from "next/link";
import { Minus, Plus, Trash2 } from "lucide-react";
import type { CartLine as Line } from "@/types/api";
import { API_URL } from "@/lib/api-url";
import { Price } from "@/components/ui/price";
import { MAX_QTY } from "../store";

type Props = {
  line: Line;
  onStep: (offerId: string, delta: number) => void;
  onRemove: (offerId: string) => void;
};

// One cart line, as in Eneba's cart popover (shopping_cart.png). Used by the
// popover and the cart page.
export function CartLine({ line, onStep, onRemove }: Props) {
  const remove = (
    <button
      type="button"
      onClick={() => onRemove(line.offerId)}
      aria-label="Remove from cart"
      className="link-hover -m-1.5 shrink-0 p-1.5 text-text-secondary"
    >
      <Trash2 className="size-4.5" />
    </button>
  );

  // A gone offer has only its id (API.md), so there is nothing else to show.
  if (line.problem === "OFFER_GONE") {
    return (
      <li className="flex items-start justify-between gap-3 py-4">
        <p className="text-sm text-red-pink">This offer is no longer available.</p>
        {remove}
      </li>
    );
  }

  const href = `/product/${line.slug}`;
  return (
    <li className="flex gap-3 py-4">
      <Link href={href} className="shrink-0">
        {/* eslint-disable-next-line @next/next/no-img-element */}
        <img src={API_URL + line.coverUrl} alt="" width={56} height={63} className="block" />
      </Link>

      <div className="min-w-0 flex-1">
        <div className="flex items-start justify-between gap-3">
          <p className="text-xs font-bold text-text-secondary">{line.seller.name}</p>
          {remove}
        </div>
        <Link href={href} className="link-hover mt-1 block text-sm font-bold">
          {line.name}
        </Link>
        <p className="mt-1 text-xs text-text-muted">Digital product</p>

        <div className="mt-2 flex items-center justify-between gap-3">
          <div className="flex items-center gap-1 text-sm">
            <QtyButton
              label="One less"
              disabled={line.quantity <= 1}
              onClick={() => onStep(line.offerId, -1)}
            >
              <Minus className="size-3.5" />
            </QtyButton>
            <span className="w-5 text-center font-semibold" aria-label="Quantity">
              {line.quantity}
            </span>
            <QtyButton
              label="One more"
              disabled={line.quantity >= MAX_QTY}
              onClick={() => onStep(line.offerId, 1)}
            >
              <Plus className="size-3.5" />
            </QtyButton>
          </div>
          <Price cents={line.lineTotalCents} className="text-base font-extrabold" />
        </div>

        {line.problem ? (
          <p className="mt-2 text-xs font-semibold text-red-pink">
            {line.problem === "OUT_OF_STOCK"
              ? "Out of stock. Remove it to check out."
              : `Only ${line.available} left. Lower the quantity to check out.`}
          </p>
        ) : (
          line.cashbackCents > 0 && (
            <p className="mt-2 text-xs text-lime">
              Cashback: <Price cents={line.cashbackCents} />
            </p>
          )
        )}
      </div>
    </li>
  );
}

type QtyProps = { label: string; disabled: boolean; onClick: () => void; children: React.ReactNode };

function QtyButton({ label, disabled, onClick, children }: QtyProps) {
  return (
    <button
      type="button"
      aria-label={label}
      disabled={disabled}
      onClick={onClick}
      className="link-hover p-1.5 disabled:opacity-40"
    >
      {children}
    </button>
  );
}
