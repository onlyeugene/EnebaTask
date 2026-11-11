import { useQueryClient } from "@tanstack/react-query";
import { useAddedModal } from "../added-modal-store";
import { addItem } from "../add-item";

// Adds one of an offer and opens the "Added to cart" box. A guest add is
// instant; a signed-in add waits for the server, so the box shows its loader
// until then. If the user closes the box during the wait, it stays closed.
export function useAddToCart() {
  const client = useQueryClient();
  const show = useAddedModal((s) => s.show);
  return async (offerId: string) => {
    show("loading");
    const outcome = await addItem(client, offerId);
    if (useAddedModal.getState().outcome === "loading") show(outcome);
  };
}
