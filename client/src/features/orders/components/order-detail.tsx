"use client";

import Link from "next/link";
import { LoaderCircle } from "lucide-react";
import type { Order } from "@/types/api";
import { ApiError } from "@/lib/client";
import { buttonStyles } from "@/components/ui/button";
import { ErrorState } from "@/components/ui/error-state";
import { Price } from "@/components/ui/price";
import { RequireAuth } from "@/features/auth/components/require-auth";
import { useOrder } from "../hooks/use-order";
import { KeyReveal } from "./key-reveal";
import { OrderStatus } from "./order-status";

// /orders/{id}, step 3 "Get your product". No Eneba screenshot of this page,
// so it uses the checkout page's panels: what happened, then each item with its keys.
export function OrderDetail({ id }: { id: string }) {
  return (
    <div className="mx-auto max-w-page px-4 pt-10 pb-10">
      <RequireAuth>
        <OrderContent id={id} />
      </RequireAuth>
    </div>
  );
}

function OrderContent({ id }: { id: string }) {
  const { data: order, isPending, error, refetch } = useOrder(id);

  if (isPending) {
    return (
      <div aria-busy className="flex justify-center py-16">
        <LoaderCircle className="size-9 animate-spin" strokeWidth={1.5} aria-hidden />
      </div>
    );
  }
  if (error) {
    // Someone else's order is a 404 too (API.md), so the same text covers both.
    if (error instanceof ApiError && error.status === 404) return <ErrorState message="Order not found." />;
    return <ErrorState message="Your order couldn't load." onRetry={() => void refetch()} />;
  }

  return (
    <div className="flex flex-col gap-6">
      <section className="bg-brand-dark p-6">
        <div className="flex flex-wrap items-center gap-3">
          <h1 className="text-3xl font-extrabold">Order {order.number}</h1>
          <OrderStatus status={order.status} />
        </div>
        <p className="mt-3 text-base text-text-secondary">{headline(order)}</p>
        <div className="mt-5 flex flex-wrap items-baseline justify-between gap-4 border-t border-divider pt-5">
          <span className="text-sm text-text-secondary">
            {new Date(order.createdAt).toLocaleString()}
            {order.payment && ` · card ending ${order.payment.last4}`}
          </span>
          <span className="text-right">
            <Price cents={order.totalCents} className="block text-3xl font-extrabold" />
            {order.status === "paid" && order.cashbackCents > 0 && (
              <span className="text-sm font-bold text-lime">
                Cashback: <Price cents={order.cashbackCents} />
              </span>
            )}
          </span>
        </div>
      </section>

      <ul className="flex flex-col gap-3">
        {order.items.map((item) => (
          <li key={item.offerId} className="bg-brand-dark p-6">
            <p className="text-sm text-text-secondary">Sold by {item.sellerName}</p>
            <Link href={`/product/${item.slug}`} className="link-hover mt-1 block text-base font-bold">
              {item.name}
            </Link>
            <p className="mt-1 text-sm text-text-secondary">
              {item.quantity} × <Price cents={item.unitPriceCents} />
            </p>
            {item.keys && item.keys.length > 0 && (
              <div className="mt-4 flex flex-col gap-2">
                <p className="text-sm font-bold">{item.keys.length === 1 ? "Your key" : "Your keys"}</p>
                {item.keys.map((k) => (
                  <KeyReveal key={k.code} code={k.code} />
                ))}
              </div>
            )}
          </li>
        ))}
      </ul>

      <div className="flex flex-wrap gap-3">
        {order.status === "paid" ? (
          <Link href="/store" className={buttonStyles.primary}>
            Keep shopping
          </Link>
        ) : (
          <Link href="/cart" className={buttonStyles.primary}>
            Back to cart
          </Link>
        )}
        <Link href="/orders" className={buttonStyles.outline}>
          All orders
        </Link>
      </div>
    </div>
  );
}

function headline(order: Order): string {
  switch (order.status) {
    case "paid":
      return "Payment received. Your keys are below; redeem them on the platform shown in the product name.";
    case "failed":
      return "The card was declined and nothing was charged. The keys went back on sale; your cart is unchanged.";
    case "expired":
      return "This order wasn't paid in time, so its keys went back on sale.";
    case "pending_payment":
      return "This order hasn't been paid yet. Check out again from your cart.";
  }
}
