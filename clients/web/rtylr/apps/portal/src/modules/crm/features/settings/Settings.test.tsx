import React from 'react';
import path from 'node:path';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { initTestI18nFrom } from '../../../../../../../test/i18nTestInstance';
import { createEffectiveAccess } from '../../../../../../../test/effectiveAccess';

vi.unmock('react-i18next');
const mocks = vi.hoisted(() => ({
  save: vi.fn(),
  pending: false,
  error: false,
  access: {} as unknown,
  settings: {
    defaultConsent: false,
    timezone: 'Europe/Paris',
    fromEmail: 'team@example.test',
    fromName: 'Test team',
    smsSenderId: '',
  },
}));
vi.mock('../../../../shared/stores/userStore', () => ({
  useUserStore: (select: (state: { user: { effectiveAccess: unknown } }) => unknown) =>
    select({ user: { effectiveAccess: mocks.access } }),
}));
vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));
vi.mock('../../shared/hooks/queries/useCrmSettings', () => ({
  useCrmSettings: () => ({
    data: mocks.settings,
    isError: mocks.error,
    refetch: vi.fn(),
  }),
  useUpdateCrmSettings: () => ({ mutate: mocks.save, isPending: mocks.pending }),
}));
import SettingsScreen from './Settings';

beforeEach(async () => {
  vi.clearAllMocks();
  mocks.pending = false;
  mocks.error = false;
  await initTestI18nFrom(path.resolve(__dirname, '../../locales'));
  mocks.access = createEffectiveAccess('crm', { isOwner: true });
});

describe('CRM settings in Portal', () => {
  it('loads existing values and saves the edited sender without duplicating team settings', () => {
    render(<SettingsScreen />);
    fireEvent.change(screen.getByLabelText('From name'), { target: { value: 'Updated team' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save changes' }));
    expect(mocks.save).toHaveBeenCalledWith(
      expect.objectContaining({
        fromName: 'Updated team',
        timezone: 'Europe/Paris',
        defaultConsent: false,
      }),
      expect.any(Object)
    );
    expect(screen.queryByRole('button', { name: 'Team' })).not.toBeInTheDocument();
  });
  it('keeps settings read only for members because updates require a business administrator', () => {
    mocks.access = createEffectiveAccess('crm');
    render(<SettingsScreen />);
    expect(screen.getByLabelText('From name')).toBeDisabled();
    expect(screen.queryByRole('button', { name: 'Save changes' })).not.toBeInTheDocument();
  });
  it('prevents saving defaults after a failed settings request', () => {
    mocks.error = true;
    render(<SettingsScreen />);
    expect(screen.getByRole('alert')).toHaveTextContent('Something went wrong');
    expect(screen.queryByRole('button', { name: 'Save changes' })).not.toBeInTheDocument();
  });
});
