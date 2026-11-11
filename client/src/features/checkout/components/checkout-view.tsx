"use client";

import Link from "next/link";
import { buttonStyles } from "@/components/ui/button";
import { RequireAuth } from "@/features/auth/components/require-auth";
import { useMe } from "@/features/auth/hooks/use-me";
import { useCart } from "@/features/cart/hooks/use-cart";
import { useGuestCartSync } from "@/features/cart/hooks/use-guest-cart-sync";
import { CartError, CartLoading, EmptyCart } from "@/features/cart/components/cart-states";
import { useResumeOrder } from "../hooks/use-resume-order";
import { CheckoutLine } from "./checkout-line";
import { PaymentForm } from "./payment-form";

// /checkout, step 2 (checkout.png): "My cart" on the left, delivery, card
// and summary on the right; stacked below 930px.
// Signed in only. A guest goes to /login?next=/checkout and comes back with
// the guest cart merged in.
export function CheckoutView() {
  return (
    <div className="mx-auto max-w-page px-4 pt-10 pb-10">
      <h1 className="sr-only">Checkout</h1>
      <RequireAuth>
        <CheckoutContent />
      </RequireAuth>
    </div>
  );
}

function CheckoutContent() {
  // If the merge at login failed, the guest lines are still saved; this sends
  // them again before the user pays, so nothing they picked is left behind.
  useGuestCartSync();
  // A refresh mid-payment: finish or drop the order this tab left held.
  useResumeOrder();
  const me = useMe();
  const { cart, loading, busy, error, retry } = useCart();

  if (loading) return <CartLoading />;
  if (error) return <CartError error={error} onRetry={retry} />;
  if (!cart || cart.items.length === 0) {
    return (
      <div>
        <EmptyCart />
        <div className="flex justify-center">
          <Link href="/store" className={buttonStyles.primary}>
            Browse the store
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="grid items-start gap-6 md:grid-cols-[1fr_360px] lg:grid-cols-[1fr_470px]">
      <section className="bg-brand-dark p-6">
        <div className="flex items-baseline justify-between gap-4">
          <h2 className="text-xl font-bold">My cart</h2>
          <Link href="/cart" className="text-sm font-bold text-yellow hover:underline">
            Edit cart
          </Link>
        </div>
        <ul className="mt-5 flex flex-col gap-3">
          {cart.items.map((l) => (
            <CheckoutLine key={l.offerId} line={l} />
          ))}
        </ul>
      </section>
      <PaymentForm cart={cart} email={me.data?.email ?? ""} cartBusy={busy} />
    </div>
  );
}
