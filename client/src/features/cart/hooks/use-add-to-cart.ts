import { useAddedModal } from "../added-modal-store";
import { useGuestCart } from "../store";

// Adds one of an offer and opens the "Added to cart" box.
// Phase 14 sends signed-in users' adds to PUT /api/cart/items instead.
export function useAddToCart() {
  const show = useAddedModal((s) => s.show);
  return (offerId: string) => show(useGuestCart.getState().add(offerId));
}
