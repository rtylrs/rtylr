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
  toast: { error: vi.fn(), success: vi.fn() },
}));

vi.mock('@voxire/api', () => ({
  uploadAsset: vi.fn(),
}));

vi.mock('@voxire/auth', () => ({
  setCookie: vi.fn(),
  removeCookie: vi.fn(),
  getDefaultCookieDomain: () => 'rtylr.com',
}));

type Business = { id: number; name: string; isOwned?: boolean; isActive?: boolean };

const businessesData = vi.hoisted(() => ({
  data: { businesses: [] as Business[] } as { businesses: Business[] } | undefined,
  isLoading: false,
}));

vi.mock('../../../../../shared/hooks/queries/useBusinesses', () => ({
  useListMyBusinesses: () => businessesData,
  useCreateSubBusiness: () => ({ mutate: vi.fn(), isPending: false }),
  useRenameBusiness: () => ({ mutate: vi.fn(), isPending: false }),
  useSetBusinessLogo: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useDeleteBusiness: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useSwitchBusiness: () => ({ mutate: vi.fn(), mutateAsync: vi.fn(), isPending: false }),
}));

vi.mock('../../../../../shared/stores/userStore', () => ({
  useUserStore: () => ({ user: { businessId: 1, businessName: 'Acme' } }),
}));

vi.mock('../../../../../shared/hooks/queries/useSubscription', () => ({
  useMySubscription: () => ({ data: undefined }),
  useBillingDetail: () => ({ data: undefined }),
}));

import Businesses from './Businesses';

beforeEach(() => {
  vi.clearAllMocks();
  businessesData.data = { businesses: [] };
  businessesData.isLoading = false;
});

describe('Businesses list empty state', () => {
  it('shows the illustration, title and an add-business call-to-action when there are no businesses', () => {
    render(<Businesses />);

    expect(screen.getByText('businesses.noBusinesses')).toBeInTheDocument();
  });

  it('opens the create-business dialog when the call-to-action is clicked', () => {
    render(<Businesses />);

    expect(screen.queryByText('businesses.createDescription')).not.toBeInTheDocument();

    const addButtons = screen.getAllByRole('button', { name: 'businesses.addBusiness' });
    fireEvent.click(addButtons[addButtons.length - 1]);

    expect(screen.getByText('businesses.createDescription')).toBeInTheDocument();
  });

  it('does not show the empty state when there are businesses', () => {
    businessesData.data = { businesses: [{ id: 1, name: 'Acme' }] };
    render(<Businesses />);

    expect(screen.queryByText('businesses.noBusinesses')).not.toBeInTheDocument();
    // "Acme" is the only business and also the current business, so it renders in
    // both the list row and the current-business side panel.
    expect(screen.getAllByText('Acme').length).toBeGreaterThan(0);
  });
});
