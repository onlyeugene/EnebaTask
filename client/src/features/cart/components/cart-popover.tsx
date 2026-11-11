"use client";

import { useEffect, useRef, useState } from "react";
import { X } from "lucide-react";
import { CartButton } from "@/components/layout/cart-button";
import { useCart } from "../hooks/use-cart";
import { useGuestCartSync } from "../hooks/use-guest-cart-sync";
import { AddedToCartModal } from "./added-to-cart-modal";
import { CartLine } from "./cart-line";
import { CartSummary } from "./cart-summary";
import { CartError, CartLoading, EmptyCart } from "./cart-states";

// The header's cart icon and the panel it opens (shopping_cart.png,
// empty_cart.png). The (shop) layout passes it into the server Header, like
// the SearchBox. It is mounted on every shop page, so it also loads the saved
// guest cart and holds the one "Added to cart" modal.
export function CartPopover() {
  useGuestCartSync();
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  const { itemCount } = useCart();

  // Close on a click outside, or on Escape.
  useEffect(() => {
    if (!open) return;
    const onDown = (e: MouseEvent) => {
      if (!ref.current?.contains(e.target as Node)) setOpen(false);
    };
    const onKey = (e: KeyboardEvent) => e.key === "Escape" && setOpen(false);
    document.addEventListener("mousedown", onDown);
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("mousedown", onDown);
      document.removeEventListener("keydown", onKey);
    };
  }, [open]);

  return (
    // No `relative` here on purpose: the panel is placed against the header's
    // row, so on a phone it stays on screen instead of hanging off the icon.
    <div ref={ref}>
      <CartButton count={itemCount} open={open} onClick={() => setOpen((o) => !o)} />
      {open && <Panel onClose={() => setOpen(false)} />}
      <AddedToCartModal />
    </div>
  );
}

function Panel({ onClose }: { onClose: () => void }) {
  const { cart, loading, busy, error, retry, step, remove, clear } = useCart();

  return (
    <div
      role="dialog"
      aria-label="My shopping cart"
      // Any link inside (product name, View cart, Checkout) leaves the page, so close.
      onClick={(e) => (e.target as HTMLElement).closest("a") && onClose()}
      className="absolute top-full right-4 z-50 w-90 max-w-[calc(100vw-32px)] border border-brand-border bg-brand-mid"
    >
      <div className="flex items-center justify-between border-b border-divider px-5 py-4">
        <h2 className="text-xl font-extrabold">My shopping cart</h2>
        <button type="button" onClick={onClose} aria-label="Close" className="link-hover -m-1 p-1">
          <X className="size-6" />
        </button>
      </div>

      {loading ? (
        <CartLoading />
      ) : error ? (
        <div className="p-5">
          <CartError error={error} onRetry={retry} onClear={clear} />
        </div>
      ) : !cart || cart.items.length === 0 ? (
        <EmptyCart />
      ) : (
        <div aria-busy={busy} className={busy ? "opacity-60" : ""}>
          {/* Long carts scroll inside the panel, so the totals stay in view. */}
          <ul className="max-h-[50vh] divide-y divide-divider overflow-y-auto border-b border-divider px-5">
            {cart.items.map((l) => (
              <CartLine key={l.offerId} line={l} onStep={step} onRemove={remove} />
            ))}
          </ul>
          <div className="p-5">
            <CartSummary cart={cart} showViewCart />
          </div>
        </div>
      )}
    </div>
  );
}
