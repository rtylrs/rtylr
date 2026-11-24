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
  tableId?: number;
  tableLabel?: string;
  fulfillmentMode: 'dine_in' | 'self_service' | 'pickup' | 'delivery';
  availableFulfillmentModes: Array<'dine_in' | 'self_service' | 'pickup' | 'delivery'>;
  currency: string;
  addItem: (item: Omit<CartItem, 'quantity' | 'notes'> & { quantity?: number }) => void;
  removeItem: (menuItemId: number) => void;
  updateQuantity: (menuItemId: number, qty: number) => void;
  setItemNote: (menuItemId: number, note: string) => void;
  setBusinessId: (id: number) => void;
  setLocationId: (id?: number) => void;
  setTable: (id?: number, label?: string) => void;
  setFulfillmentMode: (mode: 'dine_in' | 'self_service' | 'pickup' | 'delivery') => void;
  setAvailableFulfillmentModes: (
    modes: Array<'dine_in' | 'self_service' | 'pickup' | 'delivery'>
  ) => void;
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
      tableId: undefined,
      tableLabel: undefined,
      fulfillmentMode: 'dine_in',
      availableFulfillmentModes: ['dine_in', 'self_service', 'pickup'],
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

      setTable: (id, label) => {
        set({ tableId: id, tableLabel: label });
      },

      setFulfillmentMode: (mode) => {
        set({ fulfillmentMode: mode });
      },

      setAvailableFulfillmentModes: (modes) => {
        set((state) => ({
          availableFulfillmentModes: modes.length > 0 ? modes : ['dine_in'],
          fulfillmentMode:
            modes.length > 0 && !modes.includes(state.fulfillmentMode)
              ? modes[0]
              : state.fulfillmentMode,
        }));
      },

      setCurrency: (currency) => {
        set({ currency: currency || 'USD' });
      },

      clearCart: () => {
        set({
          items: [],
          businessId: 0,
          locationId: undefined,
          tableId: undefined,
          tableLabel: undefined,
          fulfillmentMode: 'dine_in',
          currency: 'USD',
        });
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
