import { Heart } from "lucide-react";
import { RIBBON } from "@/components/ui/wishlist-button";

// favorites.png's empty state: a large ribbon with an outline heart.
export function WishlistEmpty() {
  return (
    <div className="flex flex-col items-center px-4 py-16 text-center">
      <span aria-hidden className={`flex h-[84px] w-16 justify-center bg-brand-mid pt-5 ${RIBBON}`}>
        <Heart className="size-7" strokeWidth={2} />
      </span>
      <p className="mt-10 text-xl">You don&apos;t have wishlisted products yet.</p>
      <p className="mt-4 text-xl font-bold">Add products to wishlist and they will appear here.</p>
    </div>
  );
}
