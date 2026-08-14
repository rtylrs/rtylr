import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  refetch: vi.fn(),
  useGetHRConfig: vi.fn(),
  useUpsertHRConfig: vi.fn(),
}));

vi.mock('../../shared/hooks/queries/useSetup', () => ({
  useGetHRConfig: mocks.useGetHRConfig,
  useUpsertHRConfig: mocks.useUpsertHRConfig,
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

import Settings from './Settings';

describe('Settings', () => {
  beforeEach(() => {
    mocks.refetch.mockReset();
    mocks.useGetHRConfig.mockReset();
    mocks.useUpsertHRConfig.mockReturnValue({ mutate: vi.fn() });
  });

  it('shows a retryable error when the settings query cannot parse the response', () => {
    mocks.useGetHRConfig.mockReturnValue({
      data: undefined,
      isError: true,
      isFetching: false,
      isLoading: false,
      refetch: mocks.refetch,
    });

    render(<Settings />);

    expect(screen.getByRole('alert')).toHaveTextContent('Unable to load HR settings');
    expect(screen.getByRole('button', { name: 'Retry' })).toBeEnabled();
    expect(screen.queryByText('HR Settings')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'Retry' }));

    expect(mocks.refetch).toHaveBeenCalledOnce();
  });
});
