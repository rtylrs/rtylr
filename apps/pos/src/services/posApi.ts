import api from '../lib/api';

// ── Types ──────────────────────────────────────────────────────

export interface PosDashboardStats {
  totalRevenue: number;
  ordersToday: number;
  activeTables: number;
  avgOrderValue: number;
}

export interface ApiTable {
  id: string;
  label: string;
  status: 'available' | 'occupied' | 'reserved';
  seats: number;
  orderTotal?: number;
}

export interface ApiSession {
  id: string;
  staffName: string;
  openedAt: string;
  closedAt?: string;
}

export interface ApiOrderItem {
  id: string;
  name: string;
  price: number;
  qty: number;
  category: string;
  status?: string;
}

export interface ApiOrder {
  id: string;
  tableId: string;
  tableLabel: string;
  items: ApiOrderItem[];
  subtotal: number;
  tax: number;
  total: number;
  paymentMethod: 'cash' | 'card' | 'online';
  status: 'open' | 'paid' | 'voided';
  notes: string;
  createdAt: string;
}

export interface ApiLocation {
  id: string;
  name: string;
  address?: string;
}

// ── Service ────────────────────────────────────────────────────

export const posApi = {
  // ── Dashboard ─────────────────────────────────────────────────

  getDashboard: () =>
    api.get<PosDashboardStats>('/pos/dashboard').then((r) => r.data),

  // ── Tables ────────────────────────────────────────────────────

  getTables: () =>
    api.get<ApiTable[]>('/pos/tables').then((r) => r.data),

  createTable: (data: { label: string; seats: number }) =>
    api.post<ApiTable>('/pos/tables', data).then((r) => r.data),

  updateTable: (id: string, data: Partial<{ label: string; seats: number }>) =>
    api.post<ApiTable>(`/pos/tables/${id}`, data).then((r) => r.data),

  deleteTable: (id: string) =>
    api.post(`/pos/tables/${id}/delete`).then((r) => r.data),

  updateTableStatus: (id: string, status: ApiTable['status']) =>
    api.post('/pos/tables/status', { id, status }).then((r) => r.data),

  // ── Sessions ──────────────────────────────────────────────────

  getCurrentSession: () =>
    api.get<ApiSession | null>('/pos/sessions/current').then((r) => r.data),

  openSession: (data?: { staffName?: string }) =>
    api.post<ApiSession>('/pos/sessions/open', data ?? {}).then((r) => r.data),

  closeSession: (data?: { cashActual?: number }) =>
    api.post<{ success: boolean }>('/pos/sessions/close', data ?? {}).then((r) => r.data),

  // ── Orders ────────────────────────────────────────────────────

  getOrders: (params?: { status?: ApiOrder['status']; limit?: number; offset?: number }) =>
    api.get<ApiOrder[]>('/pos/orders', { params }).then((r) => r.data),

  getOrderDetail: (id: string) =>
    api.get<ApiOrder>('/pos/orders/detail', { params: { id } }).then((r) => r.data),

  createOrder: (data: {
    tableId: string;
    tableLabel: string;
    items: Omit<ApiOrderItem, 'status'>[];
    subtotal: number;
    tax: number;
    total: number;
    notes?: string;
  }) =>
    api.post<ApiOrder>('/pos/orders', data).then((r) => r.data),

  updateOrderStatus: (id: string, status: ApiOrder['status']) =>
    api.post('/pos/orders/status', { id, status }).then((r) => r.data),

  updateOrderItemStatus: (orderId: string, itemId: string, status: string) =>
    api.post('/pos/orders/item-status', { orderId, itemId, status }).then((r) => r.data),

  processPayment: (data: {
    orderId: string;
    paymentMethod: 'cash' | 'card' | 'online';
    amount: number;
  }) =>
    api.post<ApiOrder>('/pos/orders/payment', data).then((r) => r.data),

  processRefund: (data: { orderId: string; reason?: string }) =>
    api.post<{ success: boolean }>('/pos/orders/refund', data).then((r) => r.data),

  // ── Locations ─────────────────────────────────────────────────

  getLocations: () =>
    api.get<ApiLocation[]>('/pos/locations').then((r) => r.data),
};
