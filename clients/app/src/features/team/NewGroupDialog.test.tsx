import React from 'react';
import path from 'node:path';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import { initTestI18nFrom } from '../../../test/i18nTestInstance';
import { useUserStore } from '../../shared/stores/userStore';
import { NewGroupDialog } from './NewGroupDialog';

const navigateMock = vi.hoisted(() => vi.fn());
const createGroupMock = vi.hoisted(() => vi.fn());
vi.mock('@tanstack/react-router', () => ({ useNavigate: () => navigateMock }));
vi.mock('../../shared/hooks/queries/useGroups', () => ({
  useCreateGroup: () => ({ mutateAsync: createGroupMock, isPending: false }),
}));
vi.mock('../../shared/hooks/queries/usePermissions', () => ({
  usePermissionProfiles: () => ({
    data: [
      { id: '11', name: 'ERP Manager', appId: 'erp', isArchived: false },
      { id: '12', name: 'Old ERP role', appId: 'erp', isArchived: true },
      { id: '21', name: 'Cashier', appId: 'pos', isArchived: false },
    ],
  }),
}));

beforeAll(() => initTestI18nFrom(path.resolve(import.meta.dirname, '../../locales')));

beforeEach(() => {
  vi.clearAllMocks();
  createGroupMock.mockResolvedValue({ id: 6 });
  useUserStore.setState({
    user: { effectiveAccess: { subscribedApps: ['erp'] } } as never,
  });
});

describe('NewGroupDialog', () => {
  it('offers only subscribed apps and active profiles, then creates and opens the group', async () => {
    const onOpenChange = vi.fn();
    render(<NewGroupDialog open onOpenChange={onOpenChange} />);

    expect(screen.getByRole('heading', { name: 'New group' })).toBeVisible();
    expect(screen.getByText('Operations')).toBeVisible();
    expect(screen.queryByText('POS')).toBeNull();
    expect(screen.getByText('ERP Manager')).toBeVisible();
    expect(screen.queryByText('Old ERP role')).toBeNull();

    fireEvent.change(screen.getByPlaceholderText('e.g. Store Managers'), {
      target: { value: 'Operations' },
    });
    fireEvent.click(screen.getByRole('checkbox', { name: 'License Operations' }));
    fireEvent.click(screen.getByRole('checkbox', { name: 'ERP Manager' }));
    fireEvent.click(screen.getByRole('button', { name: 'New group' }));

    await waitFor(() =>
      expect(createGroupMock).toHaveBeenCalledWith(
        expect.objectContaining({
          name: 'Operations',
          color: '#6366f1',
          licensedAppIds: ['erp'],
        })
      )
    );
    expect(createGroupMock.mock.calls[0][0].permissionProfileAssignments).toEqual([
      expect.objectContaining({ appId: 'erp', permissionProfileIds: ['11'] }),
    ]);
    expect(onOpenChange).toHaveBeenCalledWith(false);
    expect(navigateMock).toHaveBeenCalledWith({ to: '/team/groups/$id', params: { id: '6' } });
  });
});
