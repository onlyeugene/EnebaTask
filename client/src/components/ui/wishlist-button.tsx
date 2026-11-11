import { Heart } from "lucide-react";

// Eneba's bookmark ribbon: a box with a V cut out of its bottom edge.
// Also used, larger, by the empty wishlist page.
export const RIBBON = "[clip-path:polygon(0_0,100%_0,100%_100%,50%_78%,0_100%)]";

type Props = {
  active: boolean;
  onClick: () => void;
  productName: string;
  // Must set the display (`flex`, or `hidden can-hover:flex`): it isn't set
  // here, because two display classes on one element fight over the order.
  className: string;
};

// The heart on product cards and the product page. Not saved: an outline
// heart, and hovering it darkens the ribbon behind it. Saved: a solid teal
// ribbon with a white filled heart. Display only: what a click does comes
// from the wishlist feature.
export function WishlistButton({ active, onClick, productName, className }: Props) {
  return (
    <button
      type="button"
      aria-pressed={active}
      aria-label={`Wishlist ${productName}`}
      title={active ? "Remove from wishlist" : "Add to wishlist"}
      onClick={onClick}
      className={`group/heart relative h-14 w-11 items-start justify-center pt-2.5 ${className}`}
    >
      <span
        aria-hidden
        className={`absolute inset-0 transition-colors ${RIBBON} ${active ? "bg-teal" : "group-hover/heart:bg-ink/50"}`}
      />
      <Heart className={`relative size-6 ${active ? "fill-white" : ""}`} strokeWidth={2} aria-hidden />
    </button>
  );
}
