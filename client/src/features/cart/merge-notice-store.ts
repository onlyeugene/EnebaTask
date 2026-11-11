import { create } from "zustand";

// How many guest lines the last merge dropped (offer gone, sold out, cart
// full), for a one-time notice. Kept in memory only, so it shows once and is
// gone after a reload. Shared state: the merge runs on the login page, and the
// notice shows on the page the user lands on next.
type MergeNotice = {
  dropped: number;
  show: (dropped: number) => void;
  close: () => void;
};

export const useMergeNotice = create<MergeNotice>()((set) => ({
  dropped: 0,
  show: (dropped) => set({ dropped }),
  close: () => set({ dropped: 0 }),
}));
