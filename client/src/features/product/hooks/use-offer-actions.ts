import { useRouter } from "next/navigation";
import { useAddedModal } from "@/features/cart/added-modal-store";
import { useAddToCart } from "@/features/cart/hooks/use-add-to-cart";
import { useGuestCart } from "@/features/cart/store";

// The two buttons on each offer.
// add: puts one in the cart and opens the "Added to cart" box.
// buyNow: puts one in the cart and goes straight to the cart page. Eneba goes
// to checkout, but ours needs sign-in (phase 14) and the cart page links on to it.
export function useOfferActions() {
  const router = useRouter();
  const add = useAddToCart();
  const show = useAddedModal((s) => s.show);

  const buyNow = (offerId: string) => {
    if (useGuestCart.getState().add(offerId) === "full") show("full");
    else router.push("/cart");
  };
  return { add, buyNow };
}
