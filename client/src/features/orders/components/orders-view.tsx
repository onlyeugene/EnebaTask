"use client";

import { RequireAuth } from "@/features/auth/components/require-auth";
import { OrderList } from "./order-list";

// /orders. Signed in only: a guest is sent to /login?next=/orders.
// pb-10: each page ends with its own 40px before the footer.
export function OrdersView() {
  return (
    <div className="mx-auto max-w-page px-4 pt-6 pb-10">
      <h1 className="text-3xl font-extrabold">My orders</h1>
      <div className="mt-6">
        <RequireAuth>
          <OrderList />
        </RequireAuth>
      </div>
    </div>
  );
}
