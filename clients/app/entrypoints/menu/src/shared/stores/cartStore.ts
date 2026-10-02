import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';

export interface CartItem {
  menuItemId: number;
  name: string;
  unitPrice: number;
  quantity: number;
  notes: string;
}

interface CartState {
  items: CartItem[];
  businessId: number;
  locationId?: number;
  addItem: (item: Omit<CartItem, 'quantity' | 'notes'>) => void;
  removeItem: (menuItemId: number) => void;
  updateQuantity: (menuItemId: number, qty: number) => void;
  setItemNote: (menuItemId: number, note: string) => void;
  setBusinessId: (id: number) => void;
  setLocationId: (id?: number) => void;
  clearCart: () => void;
  total: () => number;
  itemCount: () => number;
}

export const useCartStore = create<CartState>()(
  persist(
    (set, get) => ({
      items: [],
      businessId: 0,
      locationId: undefined,

      addItem: (item) => {
        set((state) => {
          const existing = state.items.find((i) => i.menuItemId === item.menuItemId);
          if (existing) {
            return {
              items: state.items.map((i) =>
                i.menuItemId === item.menuItemId ? { ...i, quantity: i.quantity + 1 } : i
              ),
            };
          }
          return {
            items: [...state.items, { ...item, quantity: 1, notes: '' }],
          };
        });
      },

      removeItem: (menuItemId) => {
        set((state) => ({
          items: state.items.filter((i) => i.menuItemId !== menuItemId),
        }));
      },

      updateQuantity: (menuItemId, qty) => {
        if (qty <= 0) {
          get().removeItem(menuItemId);
          return;
        }
        set((state) => ({
          items: state.items.map((i) =>
            i.menuItemId === menuItemId ? { ...i, quantity: qty } : i
          ),
        }));
      },

      setItemNote: (menuItemId, note) => {
        set((state) => ({
          items: state.items.map((i) => (i.menuItemId === menuItemId ? { ...i, notes: note } : i)),
        }));
      },

      setBusinessId: (id) => {
        set({ businessId: id });
      },

      setLocationId: (id) => {
        set({ locationId: id });
      },

      clearCart: () => {
        set({ items: [], businessId: 0, locationId: undefined });
      },

      total: () => {
        return get().items.reduce((sum, i) => sum + i.unitPrice * i.quantity, 0);
      },

      itemCount: () => {
        return get().items.reduce((sum, i) => sum + i.quantity, 0);
      },
    }),
    {
      name: 'rtylr-cart',
      storage: createJSONStorage(() => sessionStorage),
    }
  )
);
