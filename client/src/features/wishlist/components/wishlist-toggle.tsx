"use client";

import type { ProductCard } from "@/types/api";
import { WishlistButton } from "@/components/ui/wishlist-button";
import { useWishlist } from "../hooks/use-wishlist";

// The heart next to the product page title. Always shown, on touch screens too.
export function WishlistToggle({ product }: { product: ProductCard }) {
  const wishlist = useWishlist();
  return (
    <WishlistButton
      active={wishlist.has(product.id)}
      onClick={() => wishlist.toggle(product)}
      productName={product.name}
      className="flex shrink-0"
    />
  );
}
