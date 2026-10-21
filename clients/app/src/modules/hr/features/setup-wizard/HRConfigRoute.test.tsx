import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  refetch: vi.fn(),
  setConfig: vi.fn(),
  useGetHRConfig: vi.fn(),
}));

vi.mock('../../shared/hooks/queries/useSetup', () => ({
  useGetHRConfig: mocks.useGetHRConfig,
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@tanstack/react-router', () => ({
  Navigate: ({ to }: { to: string }) => <div data-testid="navigate" data-to={to} />,
  Outlet: () => <div data-testid="outlet" />,
}));

import HRConfigRoute from './HRConfigRoute';
import { useHRConfigStore } from '../../shared/stores/hrConfigStore';

describe('HRConfigRoute', () => {
  beforeEach(() => {
    mocks.refetch.mockReset();
    useHRConfigStore.getState().clearConfig();
    mocks.useGetHRConfig.mockReset();
  });

  it('keeps config failures on a retryable error instead of entering setup', () => {
    mocks.useGetHRConfig.mockReturnValue({
      data: undefined,
      isError: true,
      isFetching: false,
      isLoading: false,
      refetch: mocks.refetch,
    });

    render(<HRConfigRoute />);

    expect(screen.getByRole('alert')).toHaveTextContent('hrSetup.statusError');
    expect(screen.getByRole('button', { name: 'hrCommon.retry' })).toBeEnabled();
    expect(screen.queryByTestId('navigate')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'hrCommon.retry' }));

    expect(mocks.refetch).toHaveBeenCalledOnce();
  });

  it('renders guarded HR content for a configured tenant', () => {
    mocks.useGetHRConfig.mockReturnValue({
      data: { setupCompletedAt: 1 },
      isError: false,
      isFetching: false,
      isLoading: false,
      refetch: mocks.refetch,
    });

    render(<HRConfigRoute />);

    expect(screen.getByTestId('outlet')).toBeInTheDocument();
    expect(screen.queryByTestId('navigate')).not.toBeInTheDocument();
    expect(useHRConfigStore.getState().config).toEqual({ setupCompletedAt: 1 });
  });

  it('sends a successful response without completion to setup', () => {
    mocks.useGetHRConfig.mockReturnValue({
      data: { setupCompletedAt: undefined },
      isError: false,
      isFetching: false,
      isLoading: false,
      refetch: mocks.refetch,
    });

    render(<HRConfigRoute />);

    expect(screen.getByTestId('navigate')).toHaveAttribute('data-to', '/hr/setup');
  });
});
