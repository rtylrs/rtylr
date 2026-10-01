import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi, beforeAll } from 'vitest';
import type { EffectiveAccess } from '@voxire/types';
import Settings from './Settings';
import { initTestI18n } from '../../shared/testing/i18n';
import { ERP_THEME_STORAGE_KEY, useSettings } from '../../shared/lib/settingsStore';

const mocks = vi.hoisted(() => ({
  isAdmin: false,
  useErpSettings: vi.fn(),
  updateSettings: vi.fn(),
}));

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));
vi.mock('motion/react', () => ({
  motion: { div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div> },
}));
vi.mock('@voxire/auth', () => ({
  useUserStore: Object.assign(
    (select: (state: { effectiveAccess: EffectiveAccess }) => unknown) =>
      select({ effectiveAccess: makeAccess(mocks.isAdmin) }),
    { getState: () => ({ effectiveAccess: makeAccess(mocks.isAdmin) }) }
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
  useSettings.getState().setTheme('light');
});

describe('ERP settings permissions', () => {
  it('denies business-wide settings to a member without the permission', () => {
    render(<Settings />);
    expect(screen.queryByRole('button', { name: 'Save settings' })).not.toBeInTheDocument();
    expect(screen.getAllByRole('combobox')[0]).toBeDisabled();
  });

  it('keeps personal appearance available when business settings fail to load', () => {
    mocks.useErpSettings.mockReturnValue({
      data: undefined,
      isError: true,
      isFetching: false,
      isLoading: false,
      refetch: vi.fn(),
    });
    render(<Settings />);

    expect(screen.getByRole('alert')).toHaveTextContent('Unable to load ERP settings');
    fireEvent.click(screen.getByRole('button', { name: 'dark' }));
    expect(useSettings.getState().theme).toBe('dark');
    expect(document.documentElement).toHaveAttribute('data-theme', 'dark');
  });

  it('grants an administrator full access and links team changes to Dash', () => {
    mocks.isAdmin = true;
    render(<Settings />);

    expect(screen.getByRole('button', { name: 'Save settings' })).toBeEnabled();
    fireEvent.click(screen.getByRole('button', { name: 'Team' }));
    expect(screen.getByRole('link', { name: 'Open team management' })).toHaveAttribute(
      'href',
      expect.stringMatching(/\/team$/)
    );
    fireEvent.click(screen.getByRole('button', { name: 'General' }));
    fireEvent.click(screen.getByRole('button', { name: 'dark' }));
    expect(useSettings.getState().theme).toBe('dark');
    expect(document.documentElement).toHaveAttribute('data-theme', 'dark');
    expect(window.localStorage.getItem(ERP_THEME_STORAGE_KEY)).toBe('dark');
  });
});
