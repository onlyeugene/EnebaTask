import Link from "next/link";
import type { CartLine } from "@/types/api";
import { API_URL } from "@/lib/api-url";
import { Price } from "@/components/ui/price";

// One line of "My cart" on the checkout page (checkout.png). Read only: on
// Eneba the quantity can change here, but here the order is reserved from the
// cart on "Pay", and a retry pays that same order, so the cart must not change
// under it. "Edit cart" (step 1) is one click away.
export function CheckoutLine({ line }: { line: CartLine }) {
  if (line.problem === "OFFER_GONE") {
    return (
      <li className="bg-brand-border p-4 text-sm text-red-pink">This offer is no longer available.</li>
    );
  }

  const href = `/product/${line.slug}`;
  return (
    <li className="flex gap-4 bg-brand-border p-4">
      <Link href={href} className="shrink-0">
        {/* eslint-disable-next-line @next/next/no-img-element */}
        <img src={API_URL + line.coverUrl} alt="" width={80} height={90} className="block" />
      </Link>
      <div className="min-w-0 flex-1">
        <p className="text-sm font-bold">
          Sold by <span className="ml-1">{line.seller.name}</span>
        </p>
        <Link href={href} className="link-hover mt-1 block text-base font-bold">
          {line.name}
        </Link>
        <div className="mt-2 flex items-center justify-between gap-3 text-sm text-text-secondary">
          <span>Digital product</span>
          <span className="flex items-center gap-6">
            <span>Qty {line.quantity}</span>
            <Price cents={line.lineTotalCents} className="text-base font-bold text-white" />
          </span>
        </div>
        {line.problem ? (
          <p className="mt-2 text-xs font-semibold text-red-pink">
            {line.problem === "OUT_OF_STOCK" ? "Out of stock." : `Only ${line.available} left.`}
          </p>
        ) : (
          line.cashbackCents > 0 && (
            <p className="mt-2 text-sm text-lime">
              Cashback: <Price cents={line.cashbackCents} />
            </p>
          )
        )}
      </div>
    </li>
  );
}
