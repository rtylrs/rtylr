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
  currency: string;
  addItem: (item: Omit<CartItem, 'quantity' | 'notes'> & { quantity?: number }) => void;
  removeItem: (menuItemId: number) => void;
  updateQuantity: (menuItemId: number, qty: number) => void;
  setItemNote: (menuItemId: number, note: string) => void;
  setBusinessId: (id: number) => void;
  setLocationId: (id?: number) => void;
  setCurrency: (currency: string) => void;
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
      currency: 'USD',

      addItem: (item) => {
        const addQty = item.quantity && item.quantity > 0 ? item.quantity : 1;
        set((state) => {
          const existing = state.items.find((i) => i.menuItemId === item.menuItemId);
          if (existing) {
            return {
              items: state.items.map((i) =>
                i.menuItemId === item.menuItemId ? { ...i, quantity: i.quantity + addQty } : i
              ),
            };
          }
          return {
            items: [
              ...state.items,
              {
                menuItemId: item.menuItemId,
                name: item.name,
                unitPrice: item.unitPrice,
                quantity: addQty,
                notes: '',
              },
            ],
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

      setCurrency: (currency) => {
        set({ currency: currency || 'USD' });
      },

      clearCart: () => {
        set({ items: [], businessId: 0, locationId: undefined, currency: 'USD' });
      },

      total: () => {
        const cents = get().items.reduce(
          (sum, i) => sum + Math.round(i.unitPrice * 100) * i.quantity,
          0
        );
        return cents / 100;
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
