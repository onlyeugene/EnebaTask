import type { Metadata } from "next";
import { WishlistView } from "@/features/wishlist/components/wishlist-view";

export const metadata: Metadata = { title: "Wishlist | Eugene" };

export default function Page() {
  return <WishlistView />;
}
