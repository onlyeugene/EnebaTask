import { useRouter } from "next/navigation";
import { useQueryClient } from "@tanstack/react-query";
import { useAddedModal } from "@/features/cart/added-modal-store";
import { addItem } from "@/features/cart/add-item";
import { useAddToCart } from "@/features/cart/hooks/use-add-to-cart";

// The two buttons on each offer.
// add: puts one in the cart and opens the "Added to cart" box.
// buyNow: puts one in the cart and goes straight to the cart page. Eneba goes
// to checkout; ours goes through the cart, whose "Checkout now" leads on.
export function useOfferActions() {
  const router = useRouter();
  const client = useQueryClient();
  const add = useAddToCart();
  const show = useAddedModal((s) => s.show);

  const buyNow = async (offerId: string) => {
    const outcome = await addItem(client, offerId);
    if (outcome === "added") router.push("/cart");
    else show(outcome);
  };
  return { add, buyNow };
}
