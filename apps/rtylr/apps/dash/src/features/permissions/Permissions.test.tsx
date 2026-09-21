import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const permissionsState = vi.hoisted(() => ({
  data: undefined as
    | {
        permissions: { key: string; label: string; group: string }[];
        roles: string[];
        values: Record<string, Record<string, boolean>>;
      }
    | undefined,
  isLoading: false,
  isError: false,
  refetch: vi.fn(),
}));

vi.mock('../../shared/hooks/queries/usePermissions', () => ({
  usePermissions: () => permissionsState,
  useUpdatePermissions: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('../../shared/stores/userStore', () => ({
  useUserStore: () => ({ user: { role: 'owner' } }),
}));

import Permissions from './Permissions';

beforeEach(() => {
  vi.clearAllMocks();
  permissionsState.data = undefined;
  permissionsState.isLoading = false;
  permissionsState.isError = false;
});

describe('Permissions: failed matrix request', () => {
  it('shows a retry instead of leaving the skeleton on screen', () => {
    permissionsState.isError = true;
    render(<Permissions />);

    // Pre-fix the `isLoading || !data` guard matched a failed query forever,
    // so the skeleton was the only thing that ever rendered.
    expect(screen.getByText("Couldn't load permissions")).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Refresh' }));
    expect(permissionsState.refetch).toHaveBeenCalledTimes(1);
  });
});

describe('Permissions: matrix render', () => {
  it('renders the translated matrix rows', () => {
    permissionsState.data = {
      permissions: [{ key: 'orders.view', label: 'View Orders', group: 'Orders' }],
      roles: ['admin', 'staff'],
      values: { admin: { 'orders.view': true }, staff: { 'orders.view': false } },
    };
    render(<Permissions />);

    expect(screen.getByText('View Orders')).toBeInTheDocument();
    expect(screen.getByText('orders.view')).toBeInTheDocument();
    expect(screen.getByRole('table', { name: 'Permission Matrix' })).toBeInTheDocument();
  });
});
