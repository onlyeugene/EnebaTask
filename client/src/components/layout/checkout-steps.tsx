"use client";

import { Fragment } from "react";
import { usePathname } from "next/navigation";
import { Check } from "lucide-react";

const STEPS = ["Cart", "Payment", "Get your product"];

// Which step a page is: /cart 1, /checkout 2, an order page (/orders/{id}) 3.
// Read from the URL, so the layout needs no props from each page.
function currentStep(path: string): number {
  if (path.startsWith("/checkout")) return 2;
  if (path.startsWith("/orders/")) return 3;
  return 1;
}

// Eneba's checkout progress bar: numbered circles joined by thin lines.
// Done steps get a check, the current one is mint, later ones are dimmed.
// Below 600px only the circles and lines show.
export function CheckoutSteps() {
  const current = currentStep(usePathname());

  return (
    <ol className="flex flex-1 items-center gap-3" aria-label="Checkout progress">
      {STEPS.map((label, i) => {
        const n = i + 1;
        const state = n < current ? "done" : n === current ? "current" : "todo";
        return (
          <Fragment key={label}>
            {i > 0 && <li aria-hidden className="h-px min-w-4 flex-1 bg-text-faint" />}
            <li
              className="flex shrink-0 items-center gap-2"
              aria-current={state === "current" ? "step" : undefined}
            >
              {/* The radius tokens are reset (square corners), so the circle is set by hand. */}
              <span
                className={`flex size-6 items-center justify-center rounded-[50%] text-xs font-bold ${
                  state === "todo" ? "bg-brand-mid text-text-faint" : "bg-mint text-ink"
                }`}
              >
                {state === "done" ? <Check className="size-3.5" strokeWidth={3} aria-hidden /> : n}
              </span>
              <span
                className={`hidden text-sm sm:inline ${
                  state === "current" ? "text-mint" : state === "done" ? "text-white" : "text-text-faint"
                }`}
              >
                {label}
                {state === "done" && <span className="sr-only"> (done)</span>}
              </span>
            </li>
          </Fragment>
        );
      })}
    </ol>
  );
}
