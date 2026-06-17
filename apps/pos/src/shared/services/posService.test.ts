import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ResultAsync } from 'neverthrow';

vi.mock('./api/api', () => ({
  safe: {
    get: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: [] }), () => ({}) as any)),
    post: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: {} }), () => ({}) as any)),
    put: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: {} }), () => ({}) as any)),
    patch: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: {} }), () => ({}) as any)),
    delete: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: {} }), () => ({}) as any)),
  },
}));

import { posService } from './api/posService';
import { safe } from './api/api';

describe('posService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  // Settings
  it('getPOSSettings calls GET /pos/settings', () => {
    posService.getPOSSettings();
    expect(safe.get).toHaveBeenCalledWith('/pos/settings');
  });

  it('updatePOSSettings calls PUT /pos/settings', () => {
    const data = { business_profile: 'retail' as const, theme: 'dark' as const };
    posService.updatePOSSettings(data);
    expect(safe.put).toHaveBeenCalledWith('/pos/settings', data);
  });

  // Branches
  it('getBranches calls GET /pos/locations', () => {
    posService.getBranches();
    expect(safe.get).toHaveBeenCalledWith('/pos/locations');
  });

  it('createBranch calls POST /pos/locations', () => {
    const data = { name: 'Downtown' };
    posService.createBranch(data);
    expect(safe.post).toHaveBeenCalledWith('/pos/locations', data);
  });

  it('updateBranch calls PUT /pos/locations/:id', () => {
    posService.updateBranch(1, { name: 'Uptown' });
    expect(safe.put).toHaveBeenCalledWith('/pos/locations/1', { name: 'Uptown' });
  });

  // Floors
  it('getFloors calls GET /pos/floors with branch_id param', () => {
    posService.getFloors(2);
    expect(safe.get).toHaveBeenCalledWith('/pos/floors', { params: { branch_id: 2 } });
  });

  it('createFloor calls POST /pos/floors', () => {
    const data = { name: 'Ground' };
    posService.createFloor(data);
    expect(safe.post).toHaveBeenCalledWith('/pos/floors', data);
  });

  it('deleteFloor calls DELETE /pos/floors/:id', () => {
    posService.deleteFloor(3);
    expect(safe.delete).toHaveBeenCalledWith('/pos/floors/3');
  });

  // Terminals
  it('getTerminals calls GET /pos/terminals', () => {
    posService.getTerminals();
    expect(safe.get).toHaveBeenCalledWith('/pos/terminals');
  });

  it('createTerminal calls POST /pos/terminals with backend field names', () => {
    posService.createTerminal({ name: 'Register 1', branchId: 3, type: 'pos' });
    expect(safe.post).toHaveBeenCalledWith('/pos/terminals', {
      name: 'Register 1',
      branch_id: 3,
    });
  });

  it('updateTerminal calls PUT /pos/terminals/:id with backend field names', () => {
    posService.updateTerminal(5, {
      name: 'Front Register',
      branchId: 3,
      deviceId: 'terminal-a',
      is_active: true,
    });
    expect(safe.put).toHaveBeenCalledWith('/pos/terminals/5', {
      name: 'Front Register',
      branch_id: 3,
      device_id: 'terminal-a',
      is_active: true,
    });
  });

  it('heartbeat calls POST /pos/terminals/heartbeat with terminal id', () => {
    posService.heartbeat(9);
    expect(safe.post).toHaveBeenCalledWith('/pos/terminals/heartbeat', { terminal_id: 9 });
  });

  // Tables
  it('getTables calls GET /pos/tables', () => {
    posService.getTables();
    expect(safe.get).toHaveBeenCalledWith('/pos/tables');
  });

  it('createTable maps UI names to backend labels', () => {
    posService.createTable({ name: 'Patio 1', locationId: 2, seats: 4 });
    expect(safe.post).toHaveBeenCalledWith('/pos/tables', {
      label: 'Patio 1',
      location_id: 2,
      seats: 4,
    });
  });

  it('updateTable maps UI names to backend labels', () => {
    posService.updateTable(7, { name: 'Window 2', location_id: 3, capacity: 6 });
    expect(safe.put).toHaveBeenCalledWith('/pos/tables/7', {
      label: 'Window 2',
      location_id: 3,
      seats: 6,
    });
  });

  it('updateTableStatus calls PUT /pos/tables/:id/status', () => {
    posService.updateTableStatus(4, 'occupied');
    expect(safe.put).toHaveBeenCalledWith('/pos/tables/4/status', { status: 'occupied' });
  });

  it('deleteTable calls DELETE /pos/tables/:id', () => {
    posService.deleteTable(4);
    expect(safe.delete).toHaveBeenCalledWith('/pos/tables/4');
  });

  // Orders
  it('getOrders calls GET /pos/orders with params', () => {
    const params = { status: 'open' };
    posService.getOrders(params);
    expect(safe.get).toHaveBeenCalledWith('/pos/orders', { params });
  });

  it('getOrder calls GET /pos/orders/:id', () => {
    posService.getOrder(5);
    expect(safe.get).toHaveBeenCalledWith('/pos/orders/5');
  });

  it('createOrder calls POST /pos/orders', () => {
    const data = { items: [] };
    posService.createOrder(data);
    expect(safe.post).toHaveBeenCalledWith('/pos/orders', data);
  });

  it('updateOrderStatus calls PUT /pos/orders/:id/status', () => {
    posService.updateOrderStatus(5, 'completed');
    expect(safe.put).toHaveBeenCalledWith('/pos/orders/5/status', { status: 'completed' });
  });

  it('chargeOrder calls POST /pos/orders/:id/payment', () => {
    const data = { method: 'card' };
    posService.chargeOrder(5, data);
    expect(safe.post).toHaveBeenCalledWith('/pos/orders/5/payment', data);
  });

  it('voidOrder calls POST /pos/orders/:id/void', () => {
    posService.voidOrder(5, { reason: 'mistake' });
    expect(safe.post).toHaveBeenCalledWith('/pos/orders/5/void', { reason: 'mistake' });
  });

  it('refundOrder calls POST /pos/orders/:id/refund', () => {
    posService.refundOrder(5, { amount: 10 });
    expect(safe.post).toHaveBeenCalledWith('/pos/orders/5/refund', { amount: 10 });
  });

  it('splitOrder calls POST /pos/orders/:id/split', () => {
    posService.splitOrder(5, { parts: 2 });
    expect(safe.post).toHaveBeenCalledWith('/pos/orders/5/split', { parts: 2 });
  });

  it('getHeldOrders calls GET /pos/orders/held', () => {
    posService.getHeldOrders();
    expect(safe.get).toHaveBeenCalledWith('/pos/orders/held');
  });

  // Menu / Modifiers / Combos
  it('getMenuItems calls GET /pos/menu-items with params', () => {
    posService.getMenuItems({ category: 'burgers' });
    expect(safe.get).toHaveBeenCalledWith('/pos/menu-items', { params: { category: 'burgers' } });
  });

  it('getModifiers calls GET /pos/modifiers', () => {
    posService.getModifiers();
    expect(safe.get).toHaveBeenCalledWith('/pos/modifiers');
  });

  it('deleteModifier calls DELETE /pos/modifiers/:id', () => {
    posService.deleteModifier(2);
    expect(safe.delete).toHaveBeenCalledWith('/pos/modifiers/2');
  });

  it('getCombos calls GET /pos/combos', () => {
    posService.getCombos();
    expect(safe.get).toHaveBeenCalledWith('/pos/combos');
  });

  it('deleteCombo calls DELETE /pos/combos/:id', () => {
    posService.deleteCombo(3);
    expect(safe.delete).toHaveBeenCalledWith('/pos/combos/3');
  });

  // Sessions
  it('getSessions calls GET /pos/sessions', () => {
    posService.getSessions();
    expect(safe.get).toHaveBeenCalledWith('/pos/sessions');
  });

  it('getCurrentSession calls GET /pos/sessions/current', () => {
    posService.getCurrentSession();
    expect(safe.get).toHaveBeenCalledWith('/pos/sessions/current');
  });

  it('openSession calls POST /pos/sessions/open', () => {
    const data = { terminalId: 1 };
    posService.openSession(data);
    expect(safe.post).toHaveBeenCalledWith('/pos/sessions/open', data);
  });

  // Cash movements
  it('getCashMovements calls GET /pos/cash/movements with session_id', () => {
    posService.getCashMovements(7);
    expect(safe.get).toHaveBeenCalledWith('/pos/cash/movements', { params: { session_id: 7 } });
  });

  it('cashDrop calls POST /pos/cash/drop', () => {
    const data = { amount: 50 };
    posService.cashDrop(data);
    expect(safe.post).toHaveBeenCalledWith('/pos/cash/drop', data);
  });

  // Gift cards
  it('getGiftCards calls GET /pos/gift-cards', () => {
    posService.getGiftCards();
    expect(safe.get).toHaveBeenCalledWith('/pos/gift-cards');
  });

  it('redeemGiftCard calls POST /pos/gift-cards/:code/redeem', () => {
    posService.redeemGiftCard('ABC123', { amount: 20 });
    expect(safe.post).toHaveBeenCalledWith('/pos/gift-cards/ABC123/redeem', { amount: 20 });
  });

  it('getGiftCardBalance calls GET /pos/gift-cards/:code', () => {
    posService.getGiftCardBalance('ABC123');
    expect(safe.get).toHaveBeenCalledWith('/pos/gift-cards/ABC123');
  });

  // KDS
  it('getKDSOrders calls GET /pos/kds/orders with station param', () => {
    posService.getKDSOrders('grill');
    expect(safe.get).toHaveBeenCalledWith('/pos/kds/orders', { params: { station: 'grill' } });
  });

  it('updateKDSItemStatus calls PUT /pos/kds/items/:id', () => {
    posService.updateKDSItemStatus(10, 'ready');
    expect(safe.put).toHaveBeenCalledWith('/pos/kds/items/10', { status: 'ready' });
  });

  // Customers
  it('searchCustomers calls GET /pos/customers with q param', () => {
    posService.searchCustomers('john');
    expect(safe.get).toHaveBeenCalledWith('/pos/customers', { params: { q: 'john' } });
  });

  it('createCustomer calls POST /pos/customers', () => {
    const data = { name: 'John' };
    posService.createCustomer(data);
    expect(safe.post).toHaveBeenCalledWith('/pos/customers', data);
  });

  // Receipts
  it('emailReceipt calls POST /pos/receipts/:id/email', () => {
    posService.emailReceipt(1, 'test@example.com');
    expect(safe.post).toHaveBeenCalledWith('/pos/receipts/1/email', { email: 'test@example.com' });
  });

  // Day summary
  it('getDaySummary calls GET /pos/day-summary with date param', () => {
    posService.getDaySummary('2026-04-10');
    expect(safe.get).toHaveBeenCalledWith('/pos/day-summary', { params: { date: '2026-04-10' } });
  });

  // Manager
  it('verifyManagerPin calls POST /pos/manager/verify', () => {
    posService.verifyManagerPin('1234');
    expect(safe.post).toHaveBeenCalledWith('/pos/manager/verify', { pin: '1234' });
  });

  // Loyalty
  it('getLoyaltyConfig calls GET /pos/loyalty/config', () => {
    posService.getLoyaltyConfig();
    expect(safe.get).toHaveBeenCalledWith('/pos/loyalty/config');
  });

  it('updateLoyaltyConfig calls PUT /pos/loyalty/config', () => {
    const data = { pointsPerDollar: 5 };
    posService.updateLoyaltyConfig(data);
    expect(safe.put).toHaveBeenCalledWith('/pos/loyalty/config', data);
  });

  it('getLoyaltyBalance calls GET /pos/loyalty/balance', () => {
    posService.getLoyaltyBalance(10);
    expect(safe.get).toHaveBeenCalledWith('/pos/loyalty/balance', { params: { customer_id: 10 } });
  });

  // Stock
  it('getBranchInventory calls GET /pos/stock/inventory', () => {
    posService.getBranchInventory(2);
    expect(safe.get).toHaveBeenCalledWith('/pos/stock/inventory', { params: { branch_id: 2 } });
  });

  it('receiveStock calls POST /pos/stock/receive', () => {
    const data = { items: [] };
    posService.receiveStock(data);
    expect(safe.post).toHaveBeenCalledWith('/pos/stock/receive', data);
  });

  it('wasteStock calls POST /pos/stock/waste', () => {
    const data = { itemId: 1 };
    posService.wasteStock(data);
    expect(safe.post).toHaveBeenCalledWith('/pos/stock/waste', data);
  });

  // Audit
  it('getAuditLog calls GET /pos/audit with params', () => {
    posService.getAuditLog({ limit: 50 });
    expect(safe.get).toHaveBeenCalledWith('/pos/audit', { params: { limit: 50 } });
  });

  // POS Users
  it('getPOSUsers calls GET /pos/users with location_id', () => {
    posService.getPOSUsers(3);
    expect(safe.get).toHaveBeenCalledWith('/pos/users', { params: { location_id: 3 } });
  });

  it('deactivatePOSUser calls PATCH /pos/users/deactivate', () => {
    posService.deactivatePOSUser(5);
    expect(safe.patch).toHaveBeenCalledWith('/pos/users/deactivate', { id: 5 });
  });

  it('verifyPIN calls POST /pos/users/verify-pin with location + operator', () => {
    posService.verifyPIN(3, 42, '9999');
    expect(safe.post).toHaveBeenCalledWith('/pos/users/verify-pin', {
      location_id: 3,
      user_id: 42,
      pin: '9999',
    });
  });

  it('setPIN calls POST /pos/users/set-pin', () => {
    posService.setPIN({ userId: 42, locationId: 3, name: 'Alice', pin: '1234' });
    expect(safe.post).toHaveBeenCalledWith('/pos/users/set-pin', {
      user_id: 42,
      location_id: 3,
      name: 'Alice',
      pin: '1234',
    });
  });

  it('getLocationRoster calls GET /pos/access/location-roster', () => {
    posService.getLocationRoster(3);
    expect(safe.get).toHaveBeenCalledWith('/pos/access/location-roster', {
      params: { location_id: 3 },
    });
  });

  // Location Access
  it('getMyLocations calls GET /pos/access/my-locations', () => {
    posService.getMyLocations();
    expect(safe.get).toHaveBeenCalledWith('/pos/access/my-locations');
  });

  it('assignMemberToLocation calls POST /pos/access/assign with snake_case payload', () => {
    posService.assignMemberToLocation({
      userId: 42,
      locationId: 3,
      canManagePosUsers: true,
      canManageTerminals: false,
      canViewReports: true,
    });
    expect(safe.post).toHaveBeenCalledWith('/pos/access/assign', {
      user_id: 42,
      location_id: 3,
      can_manage_pos_users: true,
      can_manage_terminals: false,
      can_view_reports: true,
    });
  });

  it('removeMemberFromLocation calls POST /pos/access/remove', () => {
    posService.removeMemberFromLocation(8);
    expect(safe.post).toHaveBeenCalledWith('/pos/access/remove', { id: 8 });
  });
});
