import { create } from 'zustand';

export type PaymentMethod = 'cash' | 'card' | 'online';

export type TableOption = {
  id: string;
  label: string;
};

export const TABLE_OPTIONS: TableOption[] = [
  { id: 'takeaway', label: 'Takeaway' },
  { id: 'table-1',  label: 'Table 1'  },
  { id: 'table-2',  label: 'Table 2'  },
  { id: 'table-3',  label: 'Table 3'  },
  { id: 'table-4',  label: 'Table 4'  },
  { id: 'table-5',  label: 'Table 5'  },
  { id: 'table-6',  label: 'Table 6'  },
  { id: 'table-7',  label: 'Table 7'  },
  { id: 'table-8',  label: 'Table 8'  },
];

export type OrderItem = {
  id: string;
  name: string;
  price: number;
  qty: number;
  category: string;
};

export type TableStatus = 'available' | 'occupied' | 'reserved';

export type TableState = {
  id: string;
  label: string;
  status: TableStatus;
  orderTotal?: number;
  seats: number;
};

export type CompletedOrder = {
  id: string;
  tableId: string;
  tableLabel: string;
  items: OrderItem[];
  subtotal: number;
  tax: number;
  total: number;
  paymentMethod: PaymentMethod;
  status: 'open' | 'paid' | 'voided';
  notes: string;
  createdAt: Date;
};

interface POSStore {
  // Active table
  activeTableId: string;
  setActiveTableId: (id: string) => void;

  // Cart
  cartItems: OrderItem[];
  addItem: (item: Omit<OrderItem, 'qty'>) => void;
  removeItem: (id: string) => void;
  incrementQty: (id: string) => void;
  decrementQty: (id: string) => void;
  clearCart: () => void;

  // Payment
  paymentMethod: PaymentMethod;
  setPaymentMethod: (m: PaymentMethod) => void;

  // Order notes
  orderNotes: string;
  setOrderNotes: (n: string) => void;

  // Session
  staffName: string;
  sessionStart: Date;

  // Order history
  orders: CompletedOrder[];
  chargeOrder: () => void;

  // Tables
  tables: TableState[];
  setTables: (tables: TableState[]) => void;
  setTableStatus: (id: string, status: TableStatus) => void;

  // Orders (from API)
  setOrders: (orders: CompletedOrder[]) => void;
}

const INITIAL_TABLES: TableState[] = [
  { id: 'table-1', label: 'Table 1', status: 'occupied',  seats: 4, orderTotal: 48.50 },
  { id: 'table-2', label: 'Table 2', status: 'available', seats: 2 },
  { id: 'table-3', label: 'Table 3', status: 'reserved',  seats: 6 },
  { id: 'table-4', label: 'Table 4', status: 'available', seats: 4 },
  { id: 'table-5', label: 'Table 5', status: 'occupied',  seats: 8, orderTotal: 112.00 },
  { id: 'table-6', label: 'Table 6', status: 'available', seats: 2 },
  { id: 'table-7', label: 'Table 7', status: 'reserved',  seats: 4 },
  { id: 'table-8', label: 'Table 8', status: 'available', seats: 6 },
];

const SEED_ORDERS: CompletedOrder[] = [
  {
    id: 'ORD-001',
    tableId: 'table-1',
    tableLabel: 'Table 1',
    items: [
      { id: 'burger', name: 'Wagyu Burger', price: 24, qty: 2, category: 'Mains' },
      { id: 'cola',   name: 'Coca-Cola',    price: 4,  qty: 1, category: 'Drinks' },
    ],
    subtotal: 52,
    tax: 5.20,
    total: 57.20,
    paymentMethod: 'card',
    status: 'paid',
    notes: '',
    createdAt: new Date(Date.now() - 1000 * 60 * 45),
  },
  {
    id: 'ORD-002',
    tableId: 'takeaway',
    tableLabel: 'Takeaway',
    items: [
      { id: 'caesar', name: 'Caesar Salad', price: 16, qty: 1, category: 'Starters' },
    ],
    subtotal: 16,
    tax: 1.60,
    total: 17.60,
    paymentMethod: 'cash',
    status: 'paid',
    notes: 'No croutons',
    createdAt: new Date(Date.now() - 1000 * 60 * 90),
  },
];

export const usePOSStore = create<POSStore>((set, get) => ({
  activeTableId: 'takeaway',
  setActiveTableId: (id) => set({ activeTableId: id }),

  cartItems: [],
  addItem: (item) =>
    set((s) => {
      const existing = s.cartItems.find((i) => i.id === item.id);
      if (existing) {
        return {
          cartItems: s.cartItems.map((i) =>
            i.id === item.id ? { ...i, qty: i.qty + 1 } : i
          ),
        };
      }
      return { cartItems: [...s.cartItems, { ...item, qty: 1 }] };
    }),
  removeItem: (id) =>
    set((s) => ({ cartItems: s.cartItems.filter((i) => i.id !== id) })),
  incrementQty: (id) =>
    set((s) => ({
      cartItems: s.cartItems.map((i) =>
        i.id === id ? { ...i, qty: i.qty + 1 } : i
      ),
    })),
  decrementQty: (id) =>
    set((s) => ({
      cartItems: s.cartItems
        .map((i) => (i.id === id ? { ...i, qty: i.qty - 1 } : i))
        .filter((i) => i.qty > 0),
    })),
  clearCart: () => set({ cartItems: [], orderNotes: '' }),

  paymentMethod: 'card',
  setPaymentMethod: (m) => set({ paymentMethod: m }),

  orderNotes: '',
  setOrderNotes: (n) => set({ orderNotes: n }),

  staffName: 'Alex Morgan',
  sessionStart: new Date(),

  orders: SEED_ORDERS,
  chargeOrder: () => {
    const { cartItems, activeTableId, paymentMethod, orderNotes, orders } = get();
    if (cartItems.length === 0) return;
    const subtotal = cartItems.reduce((s, i) => s + i.price * i.qty, 0);
    const tax = subtotal * 0.1;
    const total = subtotal + tax;
    const tableLabel =
      TABLE_OPTIONS.find((t) => t.id === activeTableId)?.label ?? activeTableId;
    const order: CompletedOrder = {
      id: `ORD-${String(orders.length + 1).padStart(3, '0')}`,
      tableId: activeTableId,
      tableLabel,
      items: [...cartItems],
      subtotal,
      tax,
      total,
      paymentMethod,
      status: 'paid',
      notes: orderNotes,
      createdAt: new Date(),
    };
    set({ orders: [order, ...orders], cartItems: [], orderNotes: '' });
  },

  tables: INITIAL_TABLES,
  setTables: (tables) => set({ tables }),
  setTableStatus: (id, status) =>
    set((s) => ({
      tables: s.tables.map((t) => (t.id === id ? { ...t, status } : t)),
    })),

  setOrders: (orders) => set({ orders }),
}));
