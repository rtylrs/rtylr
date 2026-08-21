import { renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useAutoLogout } from './useAutoLogout';
import { useEmployeeSessionStore } from '../stores/employeeSessionStore';
import { useSettingsStore } from '../stores/settingsStore';

describe('useAutoLogout', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    useSettingsStore.setState({ autoLogoutMinutes: 1 });
    useEmployeeSessionStore.setState({
      employee: {
        id: '1',
        posUserId: '1',
        locationId: 1,
        name: 'Alice',
        role: 'cashier',
        capabilities: {
          canVoid: false,
          canRefund: false,
          canOpenDrawer: true,
          canApplyDiscount: true,
          maxDiscountPct: 10,
        },
      },
      isActive: true,
      loading: false,
      error: null,
    });
  });

  afterEach(() => {
    vi.runOnlyPendingTimers();
    vi.useRealTimers();
  });

  it('ends the employee session after the configured idle window', () => {
    renderHook(() => useAutoLogout());

    vi.advanceTimersByTime(60_000);

    expect(useEmployeeSessionStore.getState().isActive).toBe(false);
    expect(useEmployeeSessionStore.getState().employee).toBeNull();
  });

  it('resets the idle timer on user activity', () => {
    renderHook(() => useAutoLogout());

    vi.advanceTimersByTime(45_000);
    window.dispatchEvent(new KeyboardEvent('keydown'));
    vi.advanceTimersByTime(45_000);

    expect(useEmployeeSessionStore.getState().isActive).toBe(true);

    vi.advanceTimersByTime(15_000);

    expect(useEmployeeSessionStore.getState().isActive).toBe(false);
  });

  it('does not schedule logout when auto logout is disabled', () => {
    useSettingsStore.setState({ autoLogoutMinutes: 0 });

    renderHook(() => useAutoLogout());
    vi.advanceTimersByTime(24 * 60 * 60 * 1000);

    expect(useEmployeeSessionStore.getState().isActive).toBe(true);
  });
});
