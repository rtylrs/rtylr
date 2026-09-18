import { describe, it, expect, beforeEach, vi } from 'vitest';

const userClient = vi.hoisted(() => ({
  verifyPOSUserPIN: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => userClient,
}));

import { useEmployeeSessionStore } from './employeeSessionStore';

const { getState, setState } = useEmployeeSessionStore;

const verifyResponse = (overrides: Record<string, unknown> = {}): Record<string, unknown> => ({
  user_id: 42,
  pos_user_id: 7,
  location_id: 3,
  display_name: 'Alice',
  role: 'manager',
  can_void: true,
  can_refund: true,
  can_open_drawer: true,
  can_apply_discount: true,
  max_discount_pct: 100,
  ...overrides,
});

beforeEach(() => {
  setState({ employee: null, isActive: false, loading: false, error: null });
  vi.clearAllMocks();
});

describe('employeeSessionStore', () => {
  it('starts with no employee and inactive', () => {
    expect(getState().employee).toBeNull();
    expect(getState().isActive).toBe(false);
  });

  it('authenticates a roster operator and stores capabilities', async () => {
    vi.mocked(userClient.verifyPOSUserPIN).mockResolvedValueOnce(verifyResponse() as never);

    await getState().authenticatePin(3, 42, '1234');

    expect(userClient.verifyPOSUserPIN).toHaveBeenCalledWith({
      locationId: '3',
      userId: '42',
      pin: '1234',
    });
    expect(getState().isActive).toBe(true);
    expect(getState().employee).toEqual({
      id: '42',
      posUserId: '7',
      locationId: 3,
      name: 'Alice',
      role: 'manager',
      capabilities: {
        canVoid: true,
        canRefund: true,
        canOpenDrawer: true,
        canApplyDiscount: true,
        maxDiscountPct: 100,
      },
    });
  });

  it('defaults missing capability flags to false/0', async () => {
    vi.mocked(userClient.verifyPOSUserPIN).mockResolvedValueOnce(
      verifyResponse({
        role: 'cashier',
        can_void: undefined,
        can_refund: undefined,
        can_open_drawer: undefined,
        can_apply_discount: undefined,
        max_discount_pct: undefined,
      }) as never
    );

    await getState().authenticatePin(3, 42, '1234');

    expect(getState().employee?.capabilities).toEqual({
      canVoid: false,
      canRefund: false,
      canOpenDrawer: false,
      canApplyDiscount: false,
      maxDiscountPct: 0,
    });
  });

  it('sets loading true during authentication', async () => {
    let loadingDuringCall = false;
    vi.mocked(userClient.verifyPOSUserPIN).mockImplementation((() => {
      loadingDuringCall = getState().loading;
      return Promise.resolve(verifyResponse());
    }) as never);

    await getState().authenticatePin(3, 42, '1234');
    expect(loadingDuringCall).toBe(true);
  });

  it('sets a generic error on failed authentication (no cause leak)', async () => {
    vi.mocked(userClient.verifyPOSUserPIN).mockRejectedValueOnce(
      new Error('401 Invalid credentials') as never
    );

    await expect(getState().authenticatePin(3, 42, '0000')).rejects.toBeTruthy();
    expect(getState().error).toBe('Incorrect PIN');
    expect(getState().isActive).toBe(false);
    expect(getState().employee).toBeNull();
  });

  it('rejects a blank PIN without calling the API', async () => {
    await expect(getState().authenticatePin(3, 42, '   ')).rejects.toThrow(
      'Enter your employee PIN'
    );
    expect(userClient.verifyPOSUserPIN).not.toHaveBeenCalled();
    expect(getState().isActive).toBe(false);
  });

  it('rejects a malformed verify response without activating a session', async () => {
    vi.mocked(userClient.verifyPOSUserPIN).mockResolvedValueOnce(
      verifyResponse({ user_id: 0 }) as never
    );

    await expect(getState().authenticatePin(3, 42, '1234')).rejects.toThrow(
      'Employee login returned invalid employee data'
    );
    expect(getState().isActive).toBe(false);
    expect(getState().employee).toBeNull();
  });

  it('ends the session and clears the employee', () => {
    setState({ isActive: true, error: 'x' });
    getState().endSession();
    expect(getState().employee).toBeNull();
    expect(getState().isActive).toBe(false);
    expect(getState().error).toBeNull();
  });

  it('clears the error', () => {
    setState({ error: 'boom' });
    getState().clearError();
    expect(getState().error).toBeNull();
  });
});
