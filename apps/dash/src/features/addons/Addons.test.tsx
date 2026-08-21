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

const navigateMock = vi.hoisted(() => vi.fn());
vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => navigateMock,
}));

vi.mock('motion/react', () => ({
  motion: {
    div: (props: React.ComponentProps<'div'>) => <div {...props} />,
  },
  AnimatePresence: ({ children }: { children?: React.ReactNode }) => <>{children}</>,
}));

vi.mock('@tanstack/react-query', () => ({
  useMutation: () => ({ mutate: vi.fn(), mutateAsync: vi.fn(), isPending: false }),
  useQueryClient: () => ({ invalidateQueries: vi.fn() }),
}));

type AddonData = {
  id: string;
  name: string;
  description: string;
  icon: string;
  price: number;
  isFree: boolean;
  isEnabled: boolean;
};

const addonsData = vi.hoisted(() => ({
  data: [] as AddonData[],
  isLoading: false,
}));

vi.mock('../../shared/hooks/queries/useAddons', () => ({
  ADDONS_QUERY_KEY: ['addons'],
  useAddons: () => addonsData,
  enableAddon: vi.fn(),
  disableAddon: vi.fn(),
}));

vi.mock('../../shared/hooks/queries/useSubscription', () => ({
  useMySubscription: () => ({ data: undefined }),
  useBillingDetail: () => ({ data: undefined }),
  centsToDisplay: (cents: number) => `$${(cents / 100).toFixed(2)}`,
}));

import Addons from './Addons';

const makeAddon = (overrides: Partial<AddonData>): AddonData => ({
  id: 'analytics',
  name: 'Advanced Analytics',
  description: 'Deep dive into your data',
  icon: 'bar-chart-2',
  price: 0,
  isFree: true,
  isEnabled: false,
  ...overrides,
});

beforeEach(() => {
  vi.clearAllMocks();
  addonsData.data = [];
  addonsData.isLoading = false;
});

describe('Addons: search-empty state', () => {
  it('shows the illustration, title and a clear-search call-to-action when a search matches nothing', () => {
    addonsData.data = [makeAddon({})];
    render(<Addons />);

    fireEvent.change(screen.getByPlaceholderText('Search addons...'), {
      target: { value: 'no such addon' },
    });

    expect(screen.getByText('addons.emptyMatchTitle')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'common.clearSearch' })).toBeInTheDocument();
  });

  it('clears the search query when the call-to-action is clicked', () => {
    addonsData.data = [makeAddon({})];
    render(<Addons />);

    const searchInput = screen.getByPlaceholderText('Search addons...');
    fireEvent.change(searchInput, { target: { value: 'no such addon' } });
    expect(screen.getByText('addons.emptyMatchTitle')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'common.clearSearch' }));

    expect(screen.queryByText('addons.emptyMatchTitle')).not.toBeInTheDocument();
    expect(screen.getByText('Advanced Analytics')).toBeInTheDocument();
  });

  it('does not show the empty state when a search matches an addon', () => {
    addonsData.data = [makeAddon({})];
    render(<Addons />);

    fireEvent.change(screen.getByPlaceholderText('Search addons...'), {
      target: { value: 'analytics' },
    });

    expect(screen.queryByText('addons.emptyMatchTitle')).not.toBeInTheDocument();
  });
});
