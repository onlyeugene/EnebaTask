import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { CartItemInput } from "@/types/api";

// Same limits as the server (API.md, Cart): 20 lines, 1–10 per line.
// Keeping them here means a guest cart can always be quoted and merged.
export const MAX_LINES = 20;
export const MAX_QTY = 10;

// "full" when a new line would be line 21.
export type AddResult = "added" | "full";

type GuestCart = {
  // Only the choices are stored (offer + quantity). Names, prices and stock
  // come from POST /api/cart/quote on every read, so they are never stale.
  items: CartItemInput[];
  hydrated: boolean; // false until localStorage has been read
  add: (offerId: string) => AddResult;
  // +1 or -1 from the saved quantity, kept within 1–10. It reads the store, not
  // the quote on screen, so quick clicks while a quote loads all count.
  step: (offerId: string, delta: number) => void;
  remove: (offerId: string) => void;
  clear: () => void;
};

export const useGuestCart = create<GuestCart>()(
  persist(
    (set, get) => ({
      items: [],
      hydrated: false,

      // Adding an offer already in the cart raises its quantity by one
      // (up to 10), like Eneba. Lines are keyed by offerId (PLAN.md 5b).
      add: (offerId) => {
        const items = get().items;
        const line = items.find((i) => i.offerId === offerId);
        if (line) {
          set({
            items: items.map((i) =>
              i.offerId === offerId ? { ...i, quantity: Math.min(i.quantity + 1, MAX_QTY) } : i,
            ),
          });
          return "added";
        }
        if (items.length >= MAX_LINES) return "full";
        set({ items: [...items, { offerId, quantity: 1 }] });
        return "added";
      },

      step: (offerId, delta) => {
        const clamp = (q: number) => Math.min(Math.max(q, 1), MAX_QTY);
        set({
          items: get().items.map((i) =>
            i.offerId === offerId ? { ...i, quantity: clamp(i.quantity + delta) } : i,
          ),
        });
      },

      remove: (offerId) => set({ items: get().items.filter((i) => i.offerId !== offerId) }),

      clear: () => set({ items: [] }),
    }),
    {
      name: "eugene.guest-cart",
      partialize: (s) => ({ items: s.items }),
      // Not read during the first render: the server has no localStorage, so
      // reading it at once would make the first client render differ from the
      // server HTML (a hydration error). useGuestCartSync reads it after mount.
      skipHydration: true,
      onRehydrateStorage: () => () => useGuestCart.setState({ hydrated: true }),
    },
  ),
);
