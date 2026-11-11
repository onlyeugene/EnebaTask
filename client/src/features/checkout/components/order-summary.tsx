import Link from "next/link";
import { CircleHelp } from "lucide-react";
import type { Cart } from "@/types/api";
import { Button } from "@/components/ui/button";
import { Price } from "@/components/ui/price";
import { FormError } from "@/features/auth/components/form-bits";

type Props = { cart: Cart; error?: string; submitting: boolean; disabled: boolean };

// The "Summary" box from checkout.png: yellow button on top, then the count,
// service fee ("--", display only, PLAN.md section 9), total and cashback.
// Eneba's Trustpilot and newsletter boxes and "Got discount code?" are left
// out: a real brand, no newsletter, no discount codes.
export function OrderSummary({ cart, error, submitting, disabled }: Props) {
  const blocked = cart.items.some((l) => l.problem);

  return (
    <section className="bg-brand-dark p-6">
      <h2 className="text-xl font-bold">Summary</h2>

      <div className="mt-5 flex flex-col gap-3">
        <FormError message={error} />
        {blocked ? (
          <p className="text-sm text-red-pink">
            Some items can&apos;t be bought.{" "}
            <Link href="/cart" className="font-bold text-yellow hover:underline">
              Fix them in your cart
            </Link>
          </p>
        ) : (
          <Button type="submit" disabled={disabled || submitting} className="w-full text-base">
            {submitting ? "Paying…" : <>Pay <Price cents={cart.totalCents} /></>}
          </Button>
        )}
      </div>

      <div className="mt-4 flex justify-between text-sm">
        <span>{cart.itemCount === 1 ? "1 product" : `${cart.itemCount} products`}</span>
        <Price cents={cart.totalCents} className="font-bold" />
      </div>
      <div className="mt-1 flex justify-between text-sm">
        <span>Service fee</span>
        <span className="flex items-center gap-1.5">
          --
          <CircleHelp className="size-4" aria-hidden />
        </span>
      </div>

      <div className="mt-5 flex items-start justify-between gap-4 border-t border-divider pt-5">
        <span className="text-xl font-bold">Total:</span>
        <div className="text-right">
          <Price cents={cart.totalCents} className="block text-3xl font-extrabold" />
          {cart.cashbackCents > 0 && (
            <p className="text-base font-bold text-lime">
              Cashback: <Price cents={cart.cashbackCents} />
            </p>
          )}
        </div>
      </div>
    </section>
  );
}
