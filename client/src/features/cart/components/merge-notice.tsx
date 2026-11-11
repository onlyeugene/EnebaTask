"use client";

import { CircleAlert, X } from "lucide-react";
import { useMergeNotice } from "../merge-notice-store";

// The one-time notice after login when some guest lines couldn't move to the
// account cart (PLAN.md 5b: offer gone, sold out, or cart full). Rendered by
// the cart popover and the cart page, so it shows on whichever page the user
// lands on after logging in.
export function MergeNotice() {
  const { dropped, close } = useMergeNotice();
  if (dropped === 0) return null;

  return (
    <div
      role="status"
      className="fixed bottom-4 left-4 z-50 flex max-w-[calc(100vw-32px)] items-start gap-3 border border-brand-border bg-brand-mid p-4 sm:max-w-sm"
    >
      <CircleAlert className="size-5 shrink-0 text-yellow" aria-hidden />
      <p className="text-sm text-text-secondary">
        {dropped === 1 ? "1 item" : `${dropped} items`} from your guest cart couldn&apos;t be added to
        your account: sold out, no longer for sale, or your cart is full.
      </p>
      <button type="button" onClick={close} aria-label="Close" className="link-hover -m-1 p-1">
        <X className="size-4" />
      </button>
    </div>
  );
}
