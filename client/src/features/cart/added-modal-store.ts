import { create } from "zustand";
import type { AddResult } from "./store";

// Which "added to cart" message is open, if any. Shared state, because the
// buttons that add (product page, product cards on home and store) are far
// from the one <AddedToCartModal /> that the cart popover renders.
// "sold_out": the card's product sold out between page load and the click.
// "loading": a card's add is still looking up the cheapest offer.
export type AddOutcome = AddResult | "sold_out" | "error";

type AddedModal = {
  outcome: AddOutcome | "loading" | null;
  show: (outcome: AddOutcome | "loading") => void;
  close: () => void;
};

export const useAddedModal = create<AddedModal>()((set) => ({
  outcome: null,
  show: (outcome) => set({ outcome }),
  close: () => set({ outcome: null }),
}));
