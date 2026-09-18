import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import i18next from 'i18next';
import { initReactI18next } from 'react-i18next';
import enRolesPermissions from '../../../locales/en/rolesPermissions.json';

// This app's i18n bootstrap (`shared/services/i18next.ts`) loads every
// namespace file into a single 'translation' namespace, keyed by filename, so
// `t('rolesPermissions.save')` resolves as a dot-path within that one
// namespace. Reproduce that here with the real, shipped English copy so the
// assertions below check actual production strings, not raw i18n keys.
void i18next.use(initReactI18next).init({
  lng: 'en',
  fallbackLng: 'en',
  resources: { en: { translation: { rolesPermissions: enRolesPermissions } } },
  interpolation: { escapeValue: false },
});

const mocks = vi.hoisted(() => ({
  useListRoles: vi.fn(),
  useListPermissionCatalog: vi.fn(),
  useCreateRole: vi.fn(),
  useUpdateRole: vi.fn(),
}));

vi.mock('../../../shared/hooks/queries/useRoles', () => ({
  useListRoles: mocks.useListRoles,
  useListPermissionCatalog: mocks.useListPermissionCatalog,
  useCreateRole: mocks.useCreateRole,
  useUpdateRole: mocks.useUpdateRole,
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

import RolesPermissions from './RolesPermissions';

const catalog = [
  { key: 'manage_employees', label: 'Manage employees', description: '', category: 'employees' },
  { key: 'approve_time_off', label: 'Approve time off', description: '', category: 'time_off' },
];

const roles = [
  {
    id: 1,
    ownerId: 1001,
    name: 'Manager',
    description: 'Team manager',
    permissions: { approve_time_off: true },
    isSystem: true,
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
  },
];

describe('RolesPermissions', () => {
  const updateMutate = vi.fn();

  beforeEach(() => {
    updateMutate.mockReset();
    mocks.useListRoles.mockReturnValue({ data: roles, isLoading: false, isError: false });
    mocks.useListPermissionCatalog.mockReturnValue({
      data: catalog,
      isLoading: false,
      isError: false,
    });
    mocks.useCreateRole.mockReturnValue({ mutate: vi.fn(), isPending: false });
    mocks.useUpdateRole.mockReturnValue({ mutate: updateMutate, isPending: false });
  });

  it('shows a loading state while roles are fetching', () => {
    mocks.useListRoles.mockReturnValue({ data: undefined, isLoading: true, isError: false });
    render(<RolesPermissions />);
    expect(screen.getByRole('status')).toBeInTheDocument();
  });

  it('shows a retryable error state when roles fail to load', () => {
    mocks.useListRoles.mockReturnValue({ data: undefined, isLoading: false, isError: true });
    render(<RolesPermissions />);
    expect(screen.getByRole('alert')).toHaveTextContent('Unable to load roles');
  });

  it('shows the error state when only the permission catalog fails to load', () => {
    mocks.useListPermissionCatalog.mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
    });
    render(<RolesPermissions />);
    expect(screen.getByRole('alert')).toHaveTextContent('Unable to load roles');
  });

  it('renders each role with its granted/total permission count', () => {
    render(<RolesPermissions />);
    expect(screen.getByText('Manager')).toBeInTheDocument();
    expect(screen.getByText('1/2 permissions')).toBeInTheDocument();
  });

  it('toggling a permission and saving calls updateRole with the merged permissions JSON', () => {
    render(<RolesPermissions />);

    fireEvent.click(screen.getByText('Manager'));
    fireEvent.click(screen.getByRole('checkbox', { name: 'Manage employees' }));
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));

    expect(updateMutate).toHaveBeenCalledWith({
      id: 1,
      name: 'Manager',
      description: 'Team manager',
      permissions: JSON.stringify({ approve_time_off: true, manage_employees: true }),
    });
  });

  it('creating a role calls createRole with the entered name', () => {
    const createMutate = vi.fn();
    mocks.useCreateRole.mockReturnValue({ mutate: createMutate, isPending: false });

    render(<RolesPermissions />);

    fireEvent.click(screen.getByRole('button', { name: 'Add role' }));
    fireEvent.change(screen.getByPlaceholderText('Role name'), { target: { value: 'Shift Lead' } });
    fireEvent.click(screen.getByRole('button', { name: 'Create role' }));

    expect(createMutate.mock.calls[0][0]).toEqual({
      name: 'Shift Lead',
      description: '',
      permissions: '{}',
    });
  });
});
