import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  refetch: vi.fn(),
  useHrSettings: vi.fn(),
  useUpdateHrSettings: vi.fn(),
}));

vi.mock('../../shared/hooks/useHrExtras', () => ({
  useHrSettings: mocks.useHrSettings,
  useUpdateHrSettings: mocks.useUpdateHrSettings,
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('@voxire/config', () => ({
  INDUSTRIES: [],
}));

vi.mock('sonner', () => ({
  toast: { success: vi.fn() },
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

import Settings from './Settings';

describe('Settings', () => {
  beforeEach(() => {
    mocks.refetch.mockReset();
    mocks.useHrSettings.mockReset();
    mocks.useUpdateHrSettings.mockReturnValue({ mutate: vi.fn() });
  });

  it('shows a retryable error when the settings query cannot parse the response', () => {
    mocks.useHrSettings.mockReturnValue({
      data: undefined,
      isError: true,
      isFetching: false,
      isLoading: false,
      refetch: mocks.refetch,
    });

    render(<Settings />);

    expect(screen.getByRole('alert')).toHaveTextContent('errors.hrSettingsLoadFailed');
    expect(screen.getByRole('button', { name: 'common.retry' })).toBeEnabled();
    expect(screen.queryByText('HR Settings')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'common.retry' }));

    expect(mocks.refetch).toHaveBeenCalledOnce();
  });
});
