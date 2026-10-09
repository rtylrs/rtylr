import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { initTestI18n } from '../../shared/testUtils/i18nTestSetup';
import Customers from './Customers';

initTestI18n();

const mocks = vi.hoisted(() => ({
  useCustomerKPIs: vi.fn(),
  useTopCustomers: vi.fn(),
  useCustomerChurn: vi.fn(),
  useCustomerLTV: vi.fn(),
  useInsightsSettings: vi.fn(),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('../../shared/hooks/queries/useInsightsCustomers', () => ({
  useCustomerKPIs: mocks.useCustomerKPIs,
  useTopCustomers: mocks.useTopCustomers,
  useCustomerChurn: mocks.useCustomerChurn,
  useCustomerLTV: mocks.useCustomerLTV,
}));

vi.mock('../../shared/hooks/queries/useInsightsSettings', () => ({
  useInsightsSettings: mocks.useInsightsSettings,
}));

describe('Customers empty states', () => {
  const refetchTop = vi.fn();
  const refetchChurn = vi.fn();
  const refetchLtv = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    mocks.useCustomerKPIs.mockReturnValue({
      data: { total: 0, newThisMonth: 0, active: 0, churned: 0 },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    });
    mocks.useTopCustomers.mockReturnValue({
      data: { customers: [] },
      isLoading: false,
      refetch: refetchTop,
    });
    mocks.useCustomerChurn.mockReturnValue({
      data: { windows: [] },
      isLoading: false,
      refetch: refetchChurn,
    });
    mocks.useCustomerLTV.mockReturnValue({
      data: { buckets: [] },
      isLoading: false,
      refetch: refetchLtv,
    });
    mocks.useInsightsSettings.mockReturnValue({ data: { defaultCurrency: 'USD' } });
  });

  it('shows a distinct title, description and working refresh action for top customers', () => {
    render(<Customers />);

    expect(screen.getByText('No top customers yet')).toBeInTheDocument();
    expect(
      screen.getByText(
        'Your highest-spending customers will appear here once orders start coming in.'
      )
    ).toBeInTheDocument();

    fireEvent.click(screen.getAllByRole('button', { name: 'Refresh' })[0]);
    expect(refetchTop).toHaveBeenCalledTimes(1);
  });

  it('shows a distinct title, description and working refresh action for lifetime value', () => {
    render(<Customers />);

    expect(screen.getByText('No lifetime value data yet')).toBeInTheDocument();
    expect(
      screen.getByText(
        'Lifetime value segments will appear once customers build up an order history.'
      )
    ).toBeInTheDocument();

    fireEvent.click(screen.getAllByRole('button', { name: 'Refresh' })[1]);
    expect(refetchLtv).toHaveBeenCalledTimes(1);
  });

  it('shows a distinct title, description and working refresh action for churn', () => {
    render(<Customers />);

    expect(screen.getByText('No churn data yet')).toBeInTheDocument();
    expect(
      screen.getByText(
        'Churn rates will appear once there is enough customer history to measure retention.'
      )
    ).toBeInTheDocument();

    fireEvent.click(screen.getAllByRole('button', { name: 'Refresh' })[2]);
    expect(refetchChurn).toHaveBeenCalledTimes(1);
  });
});
