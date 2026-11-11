import { useQueryClient } from "@tanstack/react-query";
import type { ProductCard } from "@/types/api";
import { useAddedModal, type AddOutcome } from "@/features/cart/added-modal-store";
import { addItem } from "@/features/cart/add-item";
import { productQuery } from "./use-product";

// "Add to cart" on a product card. A card knows the product, not its offers,
// and cart lines are offers (PLAN.md 5b). So this loads the product's offers
// and adds the cheapest one in stock: the price the card shows as "From".
// staleTime 0 asks the API every time, so a just-sold-out offer isn't added.
//
// The modal opens at once with a loader while that request runs, then shows
// the result. If the user closes it during the wait, the add still happens
// but the modal doesn't pop open again.
export function useAddCheapest() {
  const client = useQueryClient();
  const show = useAddedModal((s) => s.show);

  return async (product: ProductCard) => {
    show("loading");
    let outcome: AddOutcome;
    try {
      const { offers } = await client.fetchQuery({ ...productQuery(product.slug), staleTime: 0 });
      // Offers come sorted by price with in-stock first, so the first in-stock one is cheapest.
      const offer = offers.find((o) => o.stock > 0);
      outcome = offer ? await addItem(client, offer.id) : "sold_out";
    } catch {
      outcome = "error";
    }
    if (useAddedModal.getState().outcome === "loading") show(outcome);
  };
}
