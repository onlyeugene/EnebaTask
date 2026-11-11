"use client";

import Link from "next/link";
import { Check, CircleAlert, LoaderCircle, ShoppingCart } from "lucide-react";
import { Button, buttonStyles } from "@/components/ui/button";
import { Modal } from "@/components/ui/modal";
import { useAddedModal, type AddOutcome } from "../added-modal-store";
import { useGuestCart } from "../store";

const PROBLEMS: Record<Exclude<AddOutcome, "added">, { title: string; text: string }> = {
  full: { title: "Your cart is full", text: "A cart holds up to 20 different offers." },
  sold_out: { title: "Sold out", text: "This product sold out a moment ago." },
  error: { title: "Couldn't add to cart", text: "Something went wrong. Please try again." },
};

// Eneba's "Added to cart" box (added_to_cart_modal.png). Rendered once, by the
// header's cart popover; any add button opens it through useAddedModal.
// One dialog for every state, so switching from the loader to the result only
// swaps its content, and min-h keeps the box the same size.
export function AddedToCartModal() {
  const { outcome, close } = useAddedModal();
  const count = useGuestCart((s) => s.items.reduce((n, i) => n + i.quantity, 0));
  const loading = outcome === "loading";
  const problem = outcome && outcome !== "added" && !loading ? PROBLEMS[outcome] : null;
  const title = loading ? "Adding to cart…" : (problem?.title ?? "Added to cart");

  return (
    <Modal open={outcome !== null} onClose={close} label={title}>
      <div
        aria-busy={loading}
        className="flex min-h-75 flex-col items-center justify-center px-8 py-10 text-center"
      >
        {loading ? (
          // A card's add is looking up its cheapest offer (use-add-cheapest).
          // Escape or a click outside still closes the box.
          <LoaderCircle className="size-16 animate-spin" strokeWidth={1.5} aria-hidden />
        ) : problem ? (
          <CircleAlert className="size-16 text-yellow" strokeWidth={1.5} aria-hidden />
        ) : (
          // lucide has no "cart with a tick", so a mint tick sits over the cart.
          <span className="relative">
            <ShoppingCart className="size-20" strokeWidth={1.25} aria-hidden />
            <Check className="absolute -top-2 right-0 size-8 text-mint" strokeWidth={3} aria-hidden />
          </span>
        )}

        {/* role=status: screen readers announce the change from loading to the result. */}
        <div role="status">
          <h2 className="mt-5 text-xl font-extrabold">{title}</h2>
          {!loading && (
            <p className="mt-3 text-base text-text-secondary">
              {problem ? (
                problem.text
              ) : (
                <>
                  You have <b className="text-white">{count} {count === 1 ? "item" : "items"}</b>{" "}
                  in your cart
                </>
              )}
            </p>
          )}
        </div>

        {!loading && (
          <div className="mt-8 grid w-full grid-cols-1 gap-4 xs:grid-cols-2">
            <Button onClick={close}>Continue shopping</Button>
            <Link href="/cart" onClick={close} className={buttonStyles.outline}>
              View cart
            </Link>
          </div>
        )}
      </div>
    </Modal>
  );
}
