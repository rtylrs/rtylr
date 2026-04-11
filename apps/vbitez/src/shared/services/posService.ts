import { safe } from '../lib/api';

export const posService = {
  // ── Branches (locations) ──────────────────────────────────────
  getBranches: () => safe.get('/pos/locations'),
  createBranch: (data: any) => safe.post('/pos/locations', data),
  updateBranch: (id: number, data: any) => safe.put(`/pos/locations/${id}`, data),

  // ── Floors ────────────────────────────────────────────────────
  getFloors: (branchId: number) => safe.get('/pos/floors', { params: { branch_id: branchId } }),
  createFloor: (data: any) => safe.post('/pos/floors', data),
  updateFloor: (id: number, data: any) => safe.put(`/pos/floors/${id}`, data),
  deleteFloor: (id: number) => safe.delete(`/pos/floors/${id}`),

  // ── Terminals ─────────────────────────────────────────────────
  getTerminals: () => safe.get('/pos/terminals'),
  createTerminal: (data: any) => safe.post('/pos/terminals', data),
  updateTerminal: (id: number, data: any) => safe.put(`/pos/terminals/${id}`, data),
  heartbeat: () => safe.post('/pos/terminals/heartbeat'),

  // ── Tables ────────────────────────────────────────────────────
  getTables: () => safe.get('/pos/tables'),
  createTable: (data: any) => safe.post('/pos/tables', data),
  updateTable: (id: number, data: any) => safe.put(`/pos/tables/${id}`, data),
  updateTableStatus: (id: number, status: string) => safe.put(`/pos/tables/${id}/status`, { status }),
  deleteTable: (id: number) => safe.delete(`/pos/tables/${id}`),

  // ── Orders ────────────────────────────────────────────────────
  getOrders: (params?: any) => safe.get('/pos/orders', { params }),
  getOrder: (id: number) => safe.get(`/pos/orders/${id}`),
  createOrder: (data: any) => safe.post('/pos/orders', data),
  updateOrderStatus: (id: number, status: string) => safe.put(`/pos/orders/${id}/status`, { status }),
  chargeOrder: (id: number, data: any) => safe.post(`/pos/orders/${id}/payment`, data),
  voidOrder: (id: number, data: any) => safe.post(`/pos/orders/${id}/void`, data),
  refundOrder: (id: number, data: any) => safe.post(`/pos/orders/${id}/refund`, data),
  applyDiscount: (id: number, data: any) => safe.post(`/pos/orders/${id}/discount`, data),
  splitOrder: (id: number, data: any) => safe.post(`/pos/orders/${id}/split`, data),
  holdOrder: (id: number, data: any) => safe.post(`/pos/orders/${id}/hold`, data),
  getHeldOrders: () => safe.get('/pos/orders/held'),

  // ── Menu / Products ───────────────────────────────────────────
  getMenuItems: (params?: any) => safe.get('/pos/menu-items', { params }),
  updateMenuItem: (id: number, data: any) => safe.put(`/pos/menu-items/${id}`, data),
  getModifiers: () => safe.get('/pos/modifiers'),
  createModifier: (data: any) => safe.post('/pos/modifiers', data),
  updateModifier: (id: number, data: any) => safe.put(`/pos/modifiers/${id}`, data),
  deleteModifier: (id: number) => safe.delete(`/pos/modifiers/${id}`),
  getCombos: () => safe.get('/pos/combos'),
  createCombo: (data: any) => safe.post('/pos/combos', data),
  updateCombo: (id: number, data: any) => safe.put(`/pos/combos/${id}`, data),
  deleteCombo: (id: number) => safe.delete(`/pos/combos/${id}`),

  // ── Sessions ──────────────────────────────────────────────────
  getSessions: () => safe.get('/pos/sessions'),
  getCurrentSession: () => safe.get('/pos/sessions/current'),
  openSession: (data: any) => safe.post('/pos/sessions/open', data),
  closeSession: (data: any) => safe.post('/pos/sessions/close', data),

  // ── Cash movements ────────────────────────────────────────────
  getCashMovements: (sessionId: number) => safe.get('/pos/cash/movements', { params: { session_id: sessionId } }),
  cashDrop: (data: any) => safe.post('/pos/cash/drop', data),
  cashPickup: (data: any) => safe.post('/pos/cash/pickup', data),
  cashAdjust: (data: any) => safe.post('/pos/cash/adjust', data),

  // ── Gift cards ────────────────────────────────────────────────
  getGiftCards: () => safe.get('/pos/gift-cards'),
  createGiftCard: (data: any) => safe.post('/pos/gift-cards', data),
  redeemGiftCard: (code: string, data: any) => safe.post(`/pos/gift-cards/${code}/redeem`, data),
  getGiftCardBalance: (code: string) => safe.get(`/pos/gift-cards/${code}`),

  // ── KDS ────────────────────────────────────────────────────────
  getKDSOrders: (station?: string) => safe.get('/pos/kds/orders', { params: { station } }),
  updateKDSItemStatus: (itemId: number, status: string) => safe.put(`/pos/kds/items/${itemId}`, { status }),

  // ── Customers ─────────────────────────────────────────────────
  searchCustomers: (q: string) => safe.get('/pos/customers', { params: { q } }),
  createCustomer: (data: any) => safe.post('/pos/customers', data),
  getCustomer: (id: number) => safe.get(`/pos/customers/${id}`),

  // ── Receipts ──────────────────────────────────────────────────
  getReceipts: (params?: any) => safe.get('/pos/receipts', { params }),
  emailReceipt: (id: number, email: string) => safe.post(`/pos/receipts/${id}/email`, { email }),

  // ── Day summary / Reports ─────────────────────────────────────
  getDaySummary: (date?: string) => safe.get('/pos/day-summary', { params: { date } }),

  // ── Manager override ──────────────────────────────────────────
  verifyManagerPin: (pin: string) => safe.post('/pos/manager/verify', { pin }),

  // ── Loyalty ───────────────────────────────────────────────────
  getLoyaltyConfig: () => safe.get('/pos/loyalty/config'),
  updateLoyaltyConfig: (data: any) => safe.put('/pos/loyalty/config', data),
  getLoyaltyBalance: (customerId: number) => safe.get('/pos/loyalty/balance', { params: { customer_id: customerId } }),
  earnLoyaltyPoints: (data: any) => safe.post('/pos/loyalty/earn', data),
  redeemLoyaltyPoints: (data: any) => safe.post('/pos/loyalty/redeem', data),

  // ── Stock / Inventory ─────────────────────────────────────────
  getBranchInventory: (branchId: number) => safe.get('/pos/stock/inventory', { params: { branch_id: branchId } }),
  getLowStock: (branchId: number) => safe.get('/pos/stock/low-stock', { params: { branch_id: branchId } }),
  receiveStock: (data: any) => safe.post('/pos/stock/receive', data),
  adjustStock: (data: any) => safe.post('/pos/stock/adjust', data),
  wasteStock: (data: any) => safe.post('/pos/stock/waste', data),
  transferStock: (data: any) => safe.post('/pos/stock/transfer', data),

  // ── Audit ─────────────────────────────────────────────────────
  getAuditLog: (params?: any) => safe.get('/pos/audit', { params }),

  // ── POS Users (cashiers/staff - PIN login) ────────────────────
  getPOSUsers: (locationId?: number) => safe.get('/pos/users', { params: locationId ? { location_id: locationId } : {} }),
  createPOSUser: (data: any) => safe.post('/pos/users', data),
  updatePOSUser: (data: any) => safe.post('/pos/users/update', data),
  deactivatePOSUser: (id: number) => safe.post('/pos/users/deactivate', { id }),
  verifyPIN: (pin: string) => safe.post('/pos/users/verify-pin', { pin }),

  // ── Member Location Access (which team members can access which locations) ──
  getMyLocations: () => safe.get('/pos/access/my-locations'),
  assignMemberToLocation: (data: any) => safe.post('/pos/access/assign', data),
  removeMemberFromLocation: (id: number) => safe.post('/pos/access/remove', { id }),
};
