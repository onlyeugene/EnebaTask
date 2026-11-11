"use client";

import { Info, LoaderCircle } from "lucide-react";
import { ApiError } from "@/lib/client";
import { ErrorState } from "@/components/ui/error-state";

// Eneba's empty cart (empty_cart.png).
export function EmptyCart() {
  return (
    <div className="flex flex-col items-center px-4 py-4 text-center">
      <span className="flex size-12 items-center justify-center rounded-[50%] bg-text-secondary">
        <Info className="size-7 text-brand-mid" strokeWidth={3} aria-hidden />
      </span>
      <p className="mt-4 text-xl font-extrabold text-white">Your cart is empty</p>
      <p className="mt-3 text-sm text-text-secondary">
        It seems you haven&apos;t chosen your products yet
      </p>
    </div>
  );
}

export function CartLoading() {
  return (
    <div aria-busy className="flex justify-center py-12">
      <LoaderCircle className="size-9 animate-spin" strokeWidth={1.5} aria-hidden />
    </div>
  );
}

type ErrorProps = { error: Error; onRetry: () => void; onClear: () => void };

// A 422 means the saved cart itself is broken (edited by hand, or saved by an
// older version), so asking again won't help. Offer to empty it instead.
export function CartError({ error, onRetry, onClear }: ErrorProps) {
  if (error instanceof ApiError && error.status === 422) {
    return (
      <ErrorState
        message="Your saved cart couldn't be read."
        onRetry={onClear}
        retryLabel="Empty cart"
      />
    );
  }
  return <ErrorState message="Your cart couldn't load." onRetry={onRetry} />;
}
