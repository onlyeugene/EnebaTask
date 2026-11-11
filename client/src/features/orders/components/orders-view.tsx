"use client";

import Link from "next/link";
import { buttonStyles } from "@/components/ui/button";
import { RequireAuth } from "@/features/auth/components/require-auth";

// /orders. Phase 14 only needs the page to exist behind RequireAuth (a guest
// is sent to /login?next=/orders). Phase 15 fills it with the order list.
// pb-10: each page ends with its own 40px before the footer.
export function OrdersView() {
  return (
    <div className="mx-auto max-w-page px-4 pt-6 pb-10">
      <h1 className="text-3xl font-extrabold">My orders</h1>
      <RequireAuth>
        <div className="mt-6 flex flex-col items-start gap-4">
          <p className="text-base text-text-secondary">Your orders will be listed here.</p>
          <Link href="/store" className={buttonStyles.primary}>
            Browse the store
          </Link>
        </div>
      </RequireAuth>
    </div>
  );
}
