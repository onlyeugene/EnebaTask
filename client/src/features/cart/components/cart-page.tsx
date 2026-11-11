"use client";

import Link from "next/link";
import { buttonStyles } from "@/components/ui/button";
import { useResumeOrder } from "@/features/checkout/hooks/use-resume-order";
import { useCart } from "../hooks/use-cart";
import { useGuestCartSync } from "../hooks/use-guest-cart-sync";
import { CartLine } from "./cart-line";
import { CartSummary } from "./cart-summary";
import { CartError, CartLoading, EmptyCart } from "./cart-states";
import { MergeNotice } from "./merge-notice";

// There's no screenshot of Eneba's full cart page, so this reuses the
// popover's pieces: lines on the left, totals in a box on the right.
// pb-10: each page ends with its own 40px before the footer.
// `declined`: checkout sends the user back here after a declined card.
export function CartPage({ declined = false }: { declined?: boolean }) {
  // The cart page uses the checkout layout, which has no header cart popover
  // to load the saved guest cart, so it loads it here.
  useGuestCartSync();
  // A refresh mid-payment may land here: a paid order opens its key page.
  useResumeOrder();
  const { cart, loading, busy, error, retry, step, remove, clear } = useCart();

  return (
    <div className="mx-auto max-w-page px-4 pt-6 pb-10">
      {/* Hidden on screen, kept for screen readers as the page's title. */}
      <h1 className="sr-only">My shopping cart</h1>
      <MergeNotice />
      {declined && (
        <p role="alert" className="mb-6 border border-red-dark bg-brand-darkest px-4 py-3 text-sm text-white">
          Your card was declined and nothing was charged. Your items are still in your cart, so you can try
          again.
        </p>
      )}

      {loading ? (
        <CartLoading />
      ) : error ? (
        <CartError error={error} onRetry={retry} onClear={clear} />
      ) : !cart || cart.items.length === 0 ? (
        <div className=" ">
          <EmptyCart />
          <div className="flex justify-center">
            <Link href="/" className={buttonStyles.primary}>
              Back to Homepage
            </Link>
          </div>
        </div>
      ) : (
        <div
          aria-busy={busy}
          className={`grid items-start gap-6 md:grid-cols-[1fr_360px] ${busy ? "opacity-60" : ""}`}
        >
          <ul className="divide-y divide-divider bg-brand-dark px-5">
            {cart.items.map((l) => (
              <CartLine key={l.offerId} line={l} onStep={step} onRemove={remove} />
            ))}
          </ul>
          <div className="bg-brand-mid p-5">
            <CartSummary cart={cart} />
          </div>
        </div>
      )}
    </div>
  );
}
