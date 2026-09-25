import { beforeEach, describe, expect, it } from 'vitest';
import { useSettings } from './settingsStore';

describe('useSettings persistence', () => {
  beforeEach(() => {
    localStorage.removeItem('rtylr-erp-settings');
    useSettings.getState().setTheme('dark');
    useSettings.getState().setBusinessSettings({
      currency: 'USD',
      taxRate: 8.25,
      lowStockThreshold: 10,
      defaultWarehouseId: 1,
      notifyOnLowStock: true,
      notifyOnExpiry: true,
    });
  });

  it('persists only presentation state, not tenant business settings', () => {
    useSettings.getState().setBusinessSettings({
      currency: 'EUR',
      taxRate: 20,
      lowStockThreshold: 4,
      defaultWarehouseId: 7,
      notifyOnLowStock: false,
      notifyOnExpiry: false,
    });

    const persisted = JSON.parse(localStorage.getItem('rtylr-erp-settings') ?? '{}');

    expect(persisted.state).toEqual({ theme: 'dark' });
  });
});
