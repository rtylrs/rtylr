import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('motion/react', () => ({
  motion: {
    div: (props: React.ComponentProps<'div'>) => <div {...props} />,
  },
  AnimatePresence: ({ children }: { children?: React.ReactNode }) => <>{children}</>,
}));

vi.mock('../../../../../shared/stores/businessStore', () => ({
  useBusinessStore: () => ({ businessType: '', setBusinessType: vi.fn() }),
}));

vi.mock('../../../../../shared/hooks/queries/useBusiness', () => ({
  useUpdateBusiness: () => ({ mutate: vi.fn(), isPending: false }),
}));

type IndustryType = { slug: string; label: string; icon?: string };
type IndustryGroup = { category: string; types: IndustryType[] };

const industriesData = vi.hoisted(() => ({
  data: { groups: [] as IndustryGroup[] } as { groups: IndustryGroup[] } | undefined,
  isLoading: false,
}));

vi.mock('../../../../../shared/hooks/queries/useIndustries', () => ({
  useIndustries: () => industriesData,
}));

import BusinessTypeSettings from './BusinessType';

beforeEach(() => {
  vi.clearAllMocks();
  industriesData.data = {
    groups: [
      {
        category: 'Food & Beverage',
        types: [{ slug: 'restaurant', label: 'Restaurant' }],
      },
    ],
  };
  industriesData.isLoading = false;
});

describe('BusinessType: search-empty state', () => {
  it('shows the illustration, title and a clear-search call-to-action when a search matches no industries', () => {
    render(<BusinessTypeSettings />);

    fireEvent.change(screen.getByPlaceholderText('Search industries...'), {
      target: { value: 'no such industry' },
    });

    expect(screen.getByText('settings.businessType.noIndustriesMatchTitle')).toBeInTheDocument();
    expect(
      screen.getByText('settings.businessType.noIndustriesMatchDescription')
    ).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'common.clearSearch' })).toBeInTheDocument();
  });

  it('clears the search query when the call-to-action is clicked', () => {
    render(<BusinessTypeSettings />);

    const searchInput = screen.getByPlaceholderText('Search industries...');
    fireEvent.change(searchInput, { target: { value: 'no such industry' } });
    expect(screen.getByText('settings.businessType.noIndustriesMatchTitle')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'common.clearSearch' }));

    expect(
      screen.queryByText('settings.businessType.noIndustriesMatchTitle')
    ).not.toBeInTheDocument();
    expect(screen.getByText('Restaurant')).toBeInTheDocument();
  });

  it('does not show the empty state when a search matches an industry', () => {
    render(<BusinessTypeSettings />);

    fireEvent.change(screen.getByPlaceholderText('Search industries...'), {
      target: { value: 'restaurant' },
    });

    expect(
      screen.queryByText('settings.businessType.noIndustriesMatchTitle')
    ).not.toBeInTheDocument();
  });
});
