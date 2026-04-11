import { create } from 'zustand';

export interface CartProduct {
  id: string;
  name: string;
  price: number;
  quantity: number;
  notes?: string;
}

interface CartState {
  items: CartProduct[];
  orderId: string | null;
  tableNumber: string | null;
}

interface CartActions {
  addItem: (product: Omit<CartProduct, 'quantity'>) => void;
  removeItem: (id: string) => void;
  updateQuantity: (id: string, quantity: number) => void;
  updateNotes: (id: string, notes: string) => void;
  clearCart: () => void;
  setOrderId: (id: string | null) => void;
  setTableNumber: (table: string | null) => void;
  total: () => number;
}

export const useCartStore = create<CartState & CartActions>()((set, get) => ({
  items: [],
  orderId: null,
  tableNumber: null,

  addItem: (product) => {
    set((state) => {
      const existing = state.items.find((i) => i.id === product.id);
      if (existing) {
        return {
          items: state.items.map((i) =>
            i.id === product.id ? { ...i, quantity: i.quantity + 1 } : i
          ),
        };
      }
      return { items: [...state.items, { ...product, quantity: 1 }] };
    });
  },

  removeItem: (id) => set((state) => ({ items: state.items.filter((i) => i.id !== id) })),

  updateQuantity: (id, quantity) => {
    if (quantity <= 0) {
      get().removeItem(id);
    } else {
      set((state) => ({
        items: state.items.map((i) => (i.id === id ? { ...i, quantity } : i)),
      }));
    }
  },

  updateNotes: (id, notes) =>
    set((state) => ({
      items: state.items.map((i) => (i.id === id ? { ...i, notes } : i)),
    })),

  clearCart: () => set({ items: [], orderId: null, tableNumber: null }),

  setOrderId: (orderId) => set({ orderId }),

  setTableNumber: (tableNumber) => set({ tableNumber }),

  total: () => get().items.reduce((sum, i) => sum + i.price * i.quantity, 0),
}));
