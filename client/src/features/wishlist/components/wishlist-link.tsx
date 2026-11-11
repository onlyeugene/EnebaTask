"use client";

import Link from "next/link";
import { Heart } from "lucide-react";
import { useGuestWishlistSync } from "../hooks/use-guest-wishlist-sync";
import { useWishlist } from "../hooks/use-wishlist";

// The header's heart: a link to /wishlist with a count badge when the list
// isn't empty. Mounted on every shop page, so it also loads the guest list.
export function WishlistLink() {
  useGuestWishlistSync();
  const { count } = useWishlist();

  return (
    <Link
      href="/wishlist"
      aria-label={count > 0 ? `Wishlist, ${count} ${count === 1 ? "product" : "products"}` : "Wishlist"}
      className="link-hover relative p-2.5"
    >
      <Heart className="size-7" strokeWidth={1.75} />
      {count > 0 && (
        // Same badge as the cart's (components/layout/cart-button.tsx), in teal.
        <span className="absolute top-1 right-0.5 flex h-4.5 min-w-4.5 items-center justify-center rounded-[50%] bg-teal px-1 text-2xs font-bold text-white">
          {count > 99 ? "99+" : count}
        </span>
      )}
    </Link>
  );
}
