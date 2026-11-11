import type { Cart } from "@/types/api";

// What the cart popover, cart page and modal get from useCart, whichever cart
// (guest or server) is behind it.
export type CartView = {
  cart: Cart | undefined;
  itemCount: number; // header badge
  loading: boolean;
  busy: boolean; // a change is on its way; the old cart shows dimmed
  error: Error | null;
  retry: () => void;
  step: (offerId: string, delta: number) => void;
  remove: (offerId: string) => void;
  // Empties a guest cart that can't be read. The server cart never needs it.
  clear?: () => void;
};

export const EMPTY_CART: Cart = { items: [], itemCount: 0, totalCents: 0, cashbackCents: 0 };
