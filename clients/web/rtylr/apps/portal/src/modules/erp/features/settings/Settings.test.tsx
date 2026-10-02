import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi, beforeAll } from 'vitest';
import type { EffectiveAccess } from '@voxire/types';
import Settings from './Settings';
import { initTestI18n } from '../../shared/testing/i18n';

const mocks = vi.hoisted(() => ({
  isAdmin: false,
  useErpSettings: vi.fn(),
  updateSettings: vi.fn(),
}));

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));
vi.mock('motion/react', () => ({
  motion: { div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div> },
}));
vi.mock('../../../../shared/stores/userStore', () => ({
  useUserStore: Object.assign(
    (select: (state: { user: { effectiveAccess: EffectiveAccess } }) => unknown) =>
      select({ user: { effectiveAccess: makeAccess(mocks.isAdmin) } }),
    { getState: () => ({ user: { effectiveAccess: makeAccess(mocks.isAdmin) } }) }
  ),
}));
vi.mock('../../shared/hooks/queries/useErpSettings', () => ({
  useErpSettings: mocks.useErpSettings,
  useUpdateErpSettings: () => ({ mutate: mocks.updateSettings, isPending: false }),
}));
vi.mock('../warehouses/useWarehouses', () => ({ useWarehouses: () => ({ data: [] }) }));

function makeAccess(isAdmin: boolean): EffectiveAccess {
  return {
    businessId: 'business-1',
    dashPermissions: [],
    subscribedApps: ['erp'],
    licenses: [],
    apps: [
      {
        appId: 'erp',
        licensed: true,
        hasAccess: true,
        permissions: [],
        permissionScopes: {},
        sources: ['explicit'],
      },
    ],
    isOwner: false,
    isAdmin,
    authorizationVersion: 1,
  };
}

beforeAll(async () => {
  await initTestI18n('en');
});

beforeEach(() => {
  vi.clearAllMocks();
  mocks.isAdmin = false;
  mocks.useErpSettings.mockReturnValue({
    data: {
      currency: 'USD',
      taxRate: 8.25,
      lowStockThreshold: 10,
      defaultWarehouseId: 1,
      notifyOnLowStock: true,
      notifyOnExpiry: true,
    },
    isError: false,
    isFetching: false,
    isLoading: false,
    refetch: vi.fn(),
  });
  window.localStorage.clear();
  document.documentElement.removeAttribute('data-theme');
});

describe('ERP settings permissions', () => {
  it('denies business-wide settings to a member without the permission', () => {
    render(<Settings />);
    expect(screen.queryByRole('button', { name: 'Save settings' })).not.toBeInTheDocument();
    expect(screen.getAllByRole('combobox')[0]).toBeDisabled();
  });

  it('prevents saving defaults when business settings fail to load', () => {
    mocks.useErpSettings.mockReturnValue({
      data: undefined,
      isError: true,
      isFetching: false,
      isLoading: false,
      refetch: vi.fn(),
    });
    render(<Settings />);

    expect(screen.getByRole('alert')).toHaveTextContent('Unable to load ERP settings');
    expect(screen.queryByRole('button', { name: 'Save settings' })).not.toBeInTheDocument();
  });

  it('saves business settings without duplicate Team or appearance controls', () => {
    mocks.isAdmin = true;
    render(<Settings />);
    fireEvent.click(screen.getByRole('button', { name: 'Save settings' }));
    expect(mocks.updateSettings).toHaveBeenCalledWith(
      expect.objectContaining({ currency: 'USD', taxRate: 8.25 })
    );
    expect(screen.queryByRole('button', { name: 'Team' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'dark' })).not.toBeInTheDocument();
  });
});
