import { create } from "zustand";
import { persist } from "zustand/middleware";

// Same cap as the server (API.md, Wishlist), so a guest list always fits in one merge.
export const MAX_ITEMS = 100;

// A guest's saved product. The slug is kept so the wishlist page can load the
// product (GET /api/products/{slug}); prices are never stored, so never stale.
export type GuestItem = { productId: string; slug: string; addedAt: string };

type GuestWishlist = {
  items: GuestItem[];
  hydrated: boolean; // false until localStorage has been read
  add: (productId: string, slug: string) => void;
  remove: (productId: string) => void;
};

export const useGuestWishlist = create<GuestWishlist>()(
  persist(
    (set, get) => ({
      items: [],
      hydrated: false,
      add: (productId, slug) => {
        const items = get().items;
        if (items.some((i) => i.productId === productId) || items.length >= MAX_ITEMS) return;
        set({ items: [...items, { productId, slug, addedAt: new Date().toISOString() }] });
      },
      remove: (productId) => set({ items: get().items.filter((i) => i.productId !== productId) }),
    }),
    {
      name: "eugene.guest-wishlist",
      partialize: (s) => ({ items: s.items }),
      // Read after mount, not during the first render, for the same reason as
      // the guest cart (cart/store.ts): the server HTML has no localStorage.
      skipHydration: true,
      onRehydrateStorage: () => () => useGuestWishlist.setState({ hydrated: true }),
    },
  ),
);
