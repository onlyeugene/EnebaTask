"use client";

import Link from "next/link";
import { LoaderCircle } from "lucide-react";
import { buttonStyles } from "@/components/ui/button";
import { ErrorState } from "@/components/ui/error-state";
import { Price } from "@/components/ui/price";
import { useOrders } from "../hooks/use-orders";
import { OrderStatus } from "./order-status";

// The order history, newest first. Each row opens the order page, where a
// paid order's keys are (the list itself carries no codes, API.md).
export function OrderList() {
  const { data: orders, isPending, error, refetch } = useOrders();

  if (isPending) {
    return (
      <div aria-busy className="flex justify-center py-16">
        <LoaderCircle className="size-9 animate-spin" strokeWidth={1.5} aria-hidden />
      </div>
    );
  }
  if (error) return <ErrorState message="Your orders couldn't load." onRetry={() => void refetch()} />;

  if (orders.length === 0) {
    return (
      <div className="flex flex-col items-start gap-4">
        <p className="text-base text-text-secondary">You haven&apos;t ordered anything yet.</p>
        <Link href="/store" className={buttonStyles.primary}>
          Browse the store
        </Link>
      </div>
    );
  }

  return (
    <ul className="flex flex-col gap-3">
      {orders.map((o) => (
        <li key={o.id}>
          <Link
            href={`/orders/${o.id}`}
            className="flex flex-wrap items-center justify-between gap-4 bg-brand-dark p-5 transition-colors hover:bg-brand-border"
          >
            <div className="min-w-0">
              <div className="flex flex-wrap items-center gap-3">
                <span className="text-base font-bold">{o.number}</span>
                <OrderStatus status={o.status} />
              </div>
              <p className="mt-1 truncate text-sm text-text-secondary">
                {o.items.map((i) => (i.quantity > 1 ? `${i.quantity} × ${i.name}` : i.name)).join(", ")}
              </p>
              <p className="mt-1 text-xs text-text-muted">{new Date(o.createdAt).toLocaleString()}</p>
            </div>
            <Price cents={o.totalCents} className="text-xl font-extrabold" />
          </Link>
        </li>
      ))}
    </ul>
  );
}
