import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  useGetHRConfig: vi.fn(),
  mutate: vi.fn(),
}));

vi.mock('../../shared/hooks/queries/useSetup', () => ({
  useGetHRConfig: mocks.useGetHRConfig,
  useUpsertHRConfig: () => ({ mutate: mocks.mutate, isPending: false }),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

import CompanyDirectorySetting from './CompanyDirectorySetting';

const loaded = (data: unknown) => ({ data, isLoading: false, isError: false, refetch: vi.fn() });

describe('CompanyDirectorySetting', () => {
  beforeEach(() => {
    mocks.mutate.mockReset();
  });

  it('turns the directory on with a payload carrying only the flag', () => {
    mocks.useGetHRConfig.mockReturnValue(loaded({}));
    render(<CompanyDirectorySetting />);

    const toggle = screen.getByRole('switch', { name: 'hrSetup.hrSettings.directoryTitle' });
    expect(toggle).not.toBeChecked();

    fireEvent.click(toggle);
    expect(mocks.mutate).toHaveBeenCalledWith({ directoryEnabled: true });
  });

  it('shows it off before any HR config exists', () => {
    mocks.useGetHRConfig.mockReturnValue(loaded(null));
    render(<CompanyDirectorySetting />);
    expect(screen.getByRole('switch')).not.toBeChecked();
  });

  it('turns a stored on back off', () => {
    mocks.useGetHRConfig.mockReturnValue(loaded({ directoryEnabled: true }));
    render(<CompanyDirectorySetting />);
    const toggle = screen.getByRole('switch');
    expect(toggle).toBeChecked();
    fireEvent.click(toggle);
    expect(mocks.mutate).toHaveBeenCalledWith({ directoryEnabled: false });
  });
});
