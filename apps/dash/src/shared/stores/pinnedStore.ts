import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export interface PinnedItem {
  id: string;
  label: string;
  path: string;
}

interface PinnedState {
  items: PinnedItem[];
  add: (item: PinnedItem) => void;
  remove: (id: string) => void;
}

export const usePinnedStore = create<PinnedState>()(
  persist(
    (set) => ({
      items: [],
      add: (item) =>
        set((s) => (s.items.some((i) => i.id === item.id) ? s : { items: [...s.items, item] })),
      remove: (id) => set((s) => ({ items: s.items.filter((i) => i.id !== id) })),
    }),
    { name: 'pinned-store' }
  )
);
