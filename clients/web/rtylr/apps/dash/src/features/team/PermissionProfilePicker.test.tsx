import React, { useState } from 'react';
import path from 'node:path';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { initTestI18nFrom } from '../../../../../test/i18nTestInstance';
import { PermissionProfilePicker } from './PermissionProfilePicker';

describe('PermissionProfilePicker', () => {
  it('allows independent read and write selection without discarding either profile', async () => {
    await initTestI18nFrom(path.resolve(import.meta.dirname, '../../locales'));
    const changed = vi.fn();
    function Preview() {
      const [selectedIds, setSelectedIds] = useState(['read']);
      return (
        <PermissionProfilePicker
          label="Flow"
          profiles={[
            { id: 'read', name: 'Read tasks' },
            { id: 'write', name: 'Write tasks' },
          ]}
          selectedIds={selectedIds}
          onChange={(ids) => {
            changed(ids);
            setSelectedIds(ids);
          }}
        />
      );
    }
    render(<Preview />);
    fireEvent.click(screen.getByRole('checkbox', { name: 'Write tasks' }));
    expect(changed).toHaveBeenLastCalledWith(['read', 'write']);
    expect(screen.getByRole('checkbox', { name: 'Read tasks' })).toBeChecked();
    fireEvent.click(screen.getByRole('checkbox', { name: 'Read tasks' }));
    expect(changed).toHaveBeenLastCalledWith(['write']);
    expect(screen.getByRole('checkbox', { name: 'Write tasks' })).toBeChecked();
  });

  it.each([
    ['en', 'No permission profiles available for this app.'],
    ['fr', 'Aucun profil de permissions disponible pour cette application.'],
    ['ar', 'لا تتوفر ملفات صلاحيات لهذا التطبيق.'],
  ])('renders a localized empty catalog in %s', async (language, message) => {
    await initTestI18nFrom(path.resolve(import.meta.dirname, '../../locales'), language);
    render(
      <PermissionProfilePicker label="Flow" profiles={[]} selectedIds={[]} onChange={vi.fn()} />
    );
    expect(screen.getByText(message)).toBeInTheDocument();
    expect(screen.queryByRole('checkbox')).not.toBeInTheDocument();
  });
});
