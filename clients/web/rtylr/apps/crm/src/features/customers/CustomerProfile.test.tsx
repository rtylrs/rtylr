import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createEffectiveAccess } from '../../../../../test/effectiveAccess';

const mocks = vi.hoisted(() => ({
  effectiveAccess: null as unknown,
}));

vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ id: '1' }),
  useNavigate: () => vi.fn(),
  Link: ({ children, to }: React.PropsWithChildren<{ to: string }>) => <a href={to}>{children}</a>,
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('@voxire/auth', () => {
  const state = () => ({
    user: { effectiveAccess: mocks.effectiveAccess },
    effectiveAccess: mocks.effectiveAccess,
  });
  return {
    useUserStore: Object.assign(
      (selector: (value: ReturnType<typeof state>) => unknown) => selector(state()),
      { getState: state }
    ),
  };
});

const customer = {
  id: 1,
  name: 'Jane Doe',
  email: 'jane@example.com',
  phone: '555-1234',
  createdAt: '2026-01-01',
  loyaltyPoints: 120,
  loyaltyTier: 'Bronze',
  totalSpent: 480,
  visits: 3,
  lastVisit: '2026-08-01',
  tags: [],
  notes: '',
  consent: { email: true, sms: false, marketing: true },
};

// A stable reference is required: the component's effect depends on this
// object's identity, and a mock that returns a new object on every render
// (as react-query would not, once settled) would loop forever.
const timeline = { entries: [], nextCursor: undefined };

vi.mock('../../shared/hooks/queries/useCustomers', () => ({
  useGetCustomerFull: () => ({ data: customer, isLoading: false, isError: false }),
  useCustomerTimeline: () => ({ data: timeline }),
  useUpdateCustomer: () => ({ mutate: vi.fn(), isPending: false }),
  useDeleteCustomer: () => ({ mutate: vi.fn(), isPending: false }),
  useExportCustomers: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('../../shared/hooks/queries/useSupportEngagement', () => ({
  useSupportEngagement: vi.fn(() => ({ data: null })),
}));

vi.mock('../../shared/hooks/queries/useCustomerOutstandingBalance', () => ({
  useCustomerOutstandingBalance: vi.fn(() => ({ data: null })),
}));

vi.mock('../../shared/hooks/queries/useLoyalty', () => ({
  useAwardPoints: () => ({ mutate: vi.fn(), isPending: false }),
  useRedeemPoints: () => ({ mutate: vi.fn(), isPending: false }),
  useLoyaltyConfig: () => ({ data: { programEnabled: true } }),
}));

import CustomerProfile from './CustomerProfile';

function renderProfile(): void {
  render(<CustomerProfile />);
}

beforeEach(() => {
  mocks.effectiveAccess = createEffectiveAccess('crm', { permissions: ['crm:customers:delete'] });
});

describe('CustomerProfile delete gating', () => {
  it('shows the delete button for an operator (crm:customers:manage is granted directly)', () => {
    mocks.effectiveAccess = createEffectiveAccess('crm', { permissions: ['crm:customers:delete'] });
    renderProfile();
    expect(screen.getByRole('button', { name: /delete/i })).toBeInTheDocument();
  });

  it('shows the delete button for admin (inherits crm:customers:manage)', () => {
    mocks.effectiveAccess = createEffectiveAccess('crm', { isAdmin: true });
    renderProfile();
    expect(screen.getByRole('button', { name: /delete/i })).toBeInTheDocument();
  });

  it('hides the delete button for staff, who has no CRM access', () => {
    mocks.effectiveAccess = createEffectiveAccess('crm', { permissions: [] });
    renderProfile();
    expect(screen.queryByRole('button', { name: /delete/i })).not.toBeInTheDocument();
  });

  it('hides the delete button for viewer', () => {
    mocks.effectiveAccess = createEffectiveAccess('crm', { permissions: ['crm:customers:view'] });
    renderProfile();
    expect(screen.queryByRole('button', { name: /delete/i })).not.toBeInTheDocument();
  });
});

describe('CustomerProfile tags empty state', () => {
  it('shows the empty state and focuses the tag input when the call-to-action is clicked', () => {
    mocks.effectiveAccess = createEffectiveAccess('crm', { permissions: ['crm:customers:edit'] });
    renderProfile();

    expect(screen.getByText('crm.customerProfile.noTagsTitle')).toBeInTheDocument();
    const addTag = screen.getByRole('button', { name: 'crm.customerProfile.addTag' });

    fireEvent.click(addTag);

    expect(screen.getByPlaceholderText('crm.customers.addTagPlaceholder')).toHaveFocus();
  });
});
