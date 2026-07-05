import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ResultAsync } from 'neverthrow';
import { unwrap } from '@voxire/api';

const ok = (data: unknown = { data: {} }) =>
  ResultAsync.fromPromise(Promise.resolve(data), () => ({}) as any);

vi.mock('./api/api', () => ({
  safe: {
    get: vi.fn(() => ok({ data: [] })),
    post: vi.fn(() => ok()),
    put: vi.fn(() => ok()),
    patch: vi.fn(() => ok()),
    delete: vi.fn(() => ok()),
  },
}));

import { posService } from './api/posService';
import { safe } from './api/api';

async function run(request: ReturnType<(typeof posService)[keyof typeof posService]>) {
  await unwrap(request as any);
}

describe('posService voxie client bridge', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('routes settings through the generated settings client', async () => {
    await run(posService.getPOSSettings());
    expect(safe.get).toHaveBeenCalledWith('/api/pos/settings/v1/settings', undefined);

    await run(posService.updatePOSSettings({ business_profile: 'retail', theme: 'dark' }));
    expect(safe.put).toHaveBeenCalledWith('/api/pos/settings/v1/settings', {
      businessProfile: 'retail',
      theme: 'dark',
      receiptFooter: '',
    });
  });

  it('routes locations, floors, terminals, and tables through generated paths', async () => {
    await run(posService.createBranch({ name: 'Downtown', phone: '555' }));
    expect(safe.post).toHaveBeenCalledWith('/api/pos/location/v1/locations', {
      name: 'Downtown',
      address: '',
      phone: '555',
      isActive: null,
    });

    await run(posService.getFloors(2));
    expect(safe.get).toHaveBeenCalledWith('/api/pos/floor/v1/floors', {
      params: { branch_id: '2' },
    });

    await run(posService.createTerminal({ name: 'Register 1', branchId: 3 }));
    expect(safe.post).toHaveBeenCalledWith('/api/pos/terminal/v1/terminals', {
      branchId: '3',
      name: 'Register 1',
      deviceId: '',
      isActive: true,
    });

    await run(posService.updateTableStatus(4, 'occupied'));
    expect(safe.put).toHaveBeenCalledWith('/api/pos/table/v1/tables/4', {
      locationId: null,
      label: '',
      seats: 0,
      status: 'occupied',
    });
  });

  it('builds order requests from the generated order contract', async () => {
    await run(
      posService.getOrders({
        status: 'pending',
        page: 2,
        page_size: 50,
        search: 'protein',
      })
    );
    expect(safe.get).toHaveBeenCalledWith('/api/pos/orders/v1/orders', {
      params: {
        status: 'pending',
        search: 'protein',
        page: '2',
        page_size: '50',
      },
    });

    await run(
      posService.createOrder({
        session_id: 11,
        location_id: 1,
        payment_method: 'card',
        items: [
          {
            menu_item_id: 22,
            name: 'Whey',
            quantity: 2,
            unit_price: 30,
          },
        ],
      })
    );
    expect(safe.post).toHaveBeenCalledWith('/api/pos/orders/v1/orders', {
      sessionId: '11',
      tableId: '0',
      locationId: '1',
      type: '',
      customerId: '0',
      customerName: '',
      paymentMethod: 'card',
      notes: '',
      kitchenNotes: '',
      items: [
        {
          menuItemId: '22',
          name: 'Whey',
          sku: '',
          quantity: 2,
          unitPrice: 30,
          discountPct: 0,
          taxRate: 0,
          notes: '',
        },
      ],
    });

    await run(posService.chargeOrder(5, { payment_method: 'card', amount_paid: 12 }));
    expect(safe.post).toHaveBeenCalledWith('/api/pos/orders/payments/v1/orders/5/payment', {
      orderId: '5',
      paymentMethod: 'card',
      amountPaid: 12,
    });
  });

  it('keeps local-first fallback routes where voxie has no service method yet', async () => {
    await run(posService.updateKDSItemStatus(10, 'ready'));
    expect(safe.put).toHaveBeenCalledWith('/api/pos/kds/v1/items/10', { status: 'ready' });

    await run(posService.emailReceipt(1, 'test@example.com'));
    expect(safe.post).toHaveBeenCalledWith('/api/pos/receipt/v1/receipts/1/email', {
      email: 'test@example.com',
    });

    await run(posService.getLocationRoster(3));
    expect(safe.get).toHaveBeenCalledWith('/api/pos/access/v1/location-roster', {
      params: { location_id: 3 },
    });
  });

  it('routes users and location access through generated contracts', async () => {
    await run(posService.verifyPIN(3, 42, '9999'));
    expect(safe.post).toHaveBeenCalledWith('/api/pos/user/v1/users/verify-pin', {
      locationId: '3',
      userId: '42',
      pin: '9999',
    });

    await run(
      posService.assignMemberToLocation({
        userId: 42,
        locationId: 3,
        canManagePosUsers: true,
        canManageTerminals: false,
        canViewReports: true,
      })
    );
    expect(safe.post).toHaveBeenCalledWith('/api/pos/access/v1/assign', {
      userId: '42',
      locationId: '3',
      canManagePosUsers: true,
      canManageTerminals: false,
      canViewReports: true,
    });
  });

  it('routes inventory and audit requests through generated query names', async () => {
    await run(posService.getBranchInventory(2));
    expect(safe.get).toHaveBeenCalledWith('/api/pos/stock/v1/inventory', {
      params: { branch_id: '2' },
    });

    await run(posService.receiveStock({ branch_id: 1, product_id: 2, quantity: 5 }));
    expect(safe.post).toHaveBeenCalledWith('/api/pos/stock/v1/receive', {
      branchId: '1',
      productId: '2',
      quantity: 5,
      note: '',
      employeeId: '0',
    });

    await run(posService.getAuditLog({ action: 'void', employeeId: 9 }));
    expect(safe.get).toHaveBeenCalledWith('/api/pos/audit/v1/audit', {
      params: {
        action: 'void',
        employee_id: '9',
      },
    });
  });
});
