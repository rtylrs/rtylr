import React, { useState } from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { PermissionChecklist } from './PermissionChecklist';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, options?: { query?: string }) =>
      options?.query !== undefined ? `${key}:${options.query}` : key,
  }),
}));

const permissions = [
  { key: 'erp:inventory:view', appId: 'erp', resource: 'inventory', action: 'view' },
  { key: 'erp:inventory:transfer', appId: 'erp', resource: 'inventory', action: 'edit' },
  { key: 'erp:products:edit', appId: 'erp', resource: 'products', action: 'edit' },
] as React.ComponentProps<typeof PermissionChecklist>['permissions'];

function Harness() {
  const [selected, setSelected] = useState<string[]>(['erp:inventory:view']);
  return (
    <>
      <PermissionChecklist permissions={permissions} selected={selected} onChange={setSelected} />
      <output>{selected.join(',')}</output>
    </>
  );
}

describe('PermissionChecklist', () => {
  it('filters by label or key and keeps selections made while filtered', () => {
    render(<Harness />);
    const search = screen.getByRole('searchbox', {
      name: 'common.permissionProfiles.searchPermissions',
    });

    fireEvent.change(search, { target: { value: 'transfer' } });
    expect(screen.getAllByRole('button')).toHaveLength(1);
    fireEvent.click(screen.getByRole('button', { name: /erp:inventory:transfer/ }));

    fireEvent.change(search, { target: { value: 'products' } });
    expect(screen.getByRole('button', { name: /erp:products:edit/ })).toBeVisible();

    fireEvent.change(search, { target: { value: '' } });
    expect(screen.getAllByRole('button')).toHaveLength(3);
    expect(screen.getByRole('status')).toHaveTextContent(
      'erp:inventory:view,erp:inventory:transfer'
    );
  });

  it('says so when nothing matches', () => {
    render(<Harness />);
    fireEvent.change(screen.getByRole('searchbox'), { target: { value: 'payroll' } });
    expect(screen.getByText('common.permissionProfiles.noPermissionMatches:payroll')).toBeVisible();
    expect(screen.queryAllByRole('button')).toHaveLength(0);
  });
});
