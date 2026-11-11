import Link from "next/link";
import { CircleHelp } from "lucide-react";
import type { Cart } from "@/types/api";
import { buttonStyles } from "@/components/ui/button";
import { Price } from "@/components/ui/price";

type Props = { cart: Cart; showViewCart?: boolean };

// Totals and the buttons under them, as at the bottom of Eneba's cart popover.
// The service fee is display only ("--"), as in PLAN.md section 9.
export function CartSummary({ cart, showViewCart = false }: Props) {
  // Checkout is blocked while any line can't be bought (API.md, Cart).
  const blocked = cart.items.some((l) => l.problem);

  return (
    <div>
      <div className="flex items-center justify-between text-sm text-text-secondary">
        <span>Service fee:</span>
        <span className="flex items-center gap-1.5">
          --
          <CircleHelp className="size-4" aria-hidden />
        </span>
      </div>
      <div className="mt-2 flex items-start justify-between gap-4">
        <span className="text-base font-bold">Total</span>
        <div className="text-right">
          <Price cents={cart.totalCents} className="block text-3xl font-extrabold" />
          {cart.cashbackCents > 0 && (
            <p className="text-sm font-bold text-lime">
              Cashback: <Price cents={cart.cashbackCents} />
            </p>
          )}
        </div>
      </div>

      <div className="mt-5 flex flex-col gap-3">
        {showViewCart && (
          <Link href="/cart" className={buttonStyles.primary}>
            View cart
          </Link>
        )}
        {blocked ? (
          <p className="text-center text-sm text-red-pink">
            Fix the marked items before checking out.
          </p>
        ) : (
          // Checkout needs sign-in; a guest is sent to /login and comes back.
          <Link href="/checkout" className={showViewCart ? buttonStyles.outline : buttonStyles.primary}>
            Checkout now
          </Link>
        )}
      </div>
    </div>
  );
}
