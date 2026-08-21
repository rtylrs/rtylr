import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn(), success: vi.fn(), info: vi.fn() },
}));

vi.mock('@voxire/config', () => ({
  APP_CONFIG: {},
  APP_URLS: {},
}));

vi.mock('../../shared/stores/userStore', () => ({
  useUserStore: (selector: (state: { user: { email: string } }) => unknown) =>
    selector({ user: { email: 'jane@rtylr.com' } }),
}));

type Billing = {
  isGranted: boolean;
  lineItems: unknown[];
  includedSeatLimit: number;
  totalCents: number;
  billingMonth: string;
  currentSeatCount: number;
  billableSeatCount: number;
  includedSeatCount: number;
  status: string;
};
type HistoryMonth = { month: string; status: string; totalCents: number; lineItems: unknown[] };

const subscriptionState = vi.hoisted(() => ({
  sub: { data: undefined as { status: string } | undefined, isLoading: false },
  billing: {
    data: undefined as Billing | undefined,
    isLoading: false,
    refetch: vi.fn(),
  },
  plans: { data: [] as unknown[], isLoading: false },
  history: {
    data: [] as HistoryMonth[],
    isLoading: false,
    refetch: vi.fn(),
  },
}));

vi.mock('../../shared/hooks/queries/useSubscription', () => ({
  useMySubscription: () => subscriptionState.sub,
  useBillingDetail: () => subscriptionState.billing,
  useListPlans: () => subscriptionState.plans,
  useBillingHistory: () => subscriptionState.history,
  useCheckout: () => ({ mutate: vi.fn(), isPending: false }),
  centsToDisplay: (cents: number) => `$${(cents / 100).toFixed(2)}`,
  trialDaysLeft: () => 0,
}));

import Billing from './Billing';

const makeBilling = (overrides: Partial<Billing>): Billing => ({
  isGranted: false,
  lineItems: [],
  includedSeatLimit: 10,
  totalCents: 0,
  billingMonth: new Date().toISOString(),
  currentSeatCount: 0,
  billableSeatCount: 0,
  includedSeatCount: 0,
  status: 'active',
  ...overrides,
});

beforeEach(() => {
  vi.clearAllMocks();
  subscriptionState.sub = { data: undefined, isLoading: false };
  subscriptionState.billing = { data: makeBilling({}), isLoading: false, refetch: vi.fn() };
  subscriptionState.plans = { data: [], isLoading: false };
  subscriptionState.history = { data: [], isLoading: false, refetch: vi.fn() };
});

describe('Billing: current-charges empty state', () => {
  it('shows the illustration, title and a refresh call-to-action when there is no billable usage', () => {
    render(<Billing />);

    expect(screen.getByText('billing.emptyNoUsageTitle')).toBeInTheDocument();
  });

  it('refetches billing detail when the current-charges refresh action is clicked', () => {
    render(<Billing />);

    const refreshButtons = screen.getAllByRole('button', { name: 'common.refresh' });
    fireEvent.click(refreshButtons[0]);

    expect(subscriptionState.billing.refetch).toHaveBeenCalledTimes(1);
  });

  it('does not show the current-charges empty state when there is billable usage', () => {
    subscriptionState.billing = {
      data: makeBilling({
        lineItems: [{ type: 'seat', seatCount: 1 }],
      }),
      isLoading: false,
      refetch: vi.fn(),
    };
    render(<Billing />);

    expect(screen.queryByText('billing.emptyNoUsageTitle')).not.toBeInTheDocument();
  });
});

describe('Billing: invoice-history empty state', () => {
  it('shows the illustration, title, description and a refresh call-to-action when there is no invoice history', () => {
    render(<Billing />);

    expect(screen.getByText('billing.emptyNoInvoicesTitle')).toBeInTheDocument();
    expect(screen.getByText('billing.emptyNoInvoicesDescription')).toBeInTheDocument();
  });

  it('refetches billing history when the invoice-history refresh action is clicked', () => {
    render(<Billing />);

    const refreshButtons = screen.getAllByRole('button', { name: 'common.refresh' });
    fireEvent.click(refreshButtons[refreshButtons.length - 1]);

    expect(subscriptionState.history.refetch).toHaveBeenCalledTimes(1);
  });

  it('does not show the invoice-history empty state when there is history', () => {
    subscriptionState.history = {
      data: [{ month: '2026-07', status: 'paid', totalCents: 1000, lineItems: [] }],
      isLoading: false,
      refetch: vi.fn(),
    };
    render(<Billing />);

    expect(screen.queryByText('billing.emptyNoInvoicesTitle')).not.toBeInTheDocument();
  });
});
