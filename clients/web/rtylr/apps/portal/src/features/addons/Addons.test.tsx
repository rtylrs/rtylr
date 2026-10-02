import path from 'node:path';
import i18next from 'i18next';
import { initTestI18nFrom } from '../../../../../test/i18nTestInstance';
import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, beforeAll } from 'vitest';

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

beforeAll(() => initTestI18nFrom(path.resolve(import.meta.dirname, '../../locales')));

beforeEach(() => {
  vi.clearAllMocks();
  addonsData.data = [];
  addonsData.isLoading = false;
});

describe('Addons: search-empty state', () => {
  it('shows the illustration, title and a clear-search call-to-action when a search matches nothing', () => {
    addonsData.data = [makeAddon({})];
    render(<Addons />);

    fireEvent.change(screen.getByPlaceholderText(i18next.t('common.management.searchAddons')), {
      target: { value: 'no such addon' },
    });

    expect(screen.getByText(i18next.t('addons.emptyMatchTitle'))).toBeInTheDocument();
    expect(
      screen.getByRole('button', { name: i18next.t('common.clearSearch') })
    ).toBeInTheDocument();
  });

  it('clears the search query when the call-to-action is clicked', () => {
    addonsData.data = [makeAddon({})];
    render(<Addons />);

    const searchInput = screen.getByPlaceholderText(i18next.t('common.management.searchAddons'));
    fireEvent.change(searchInput, { target: { value: 'no such addon' } });
    expect(screen.getByText(i18next.t('addons.emptyMatchTitle'))).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: i18next.t('common.clearSearch') }));

    expect(screen.queryByText(i18next.t('addons.emptyMatchTitle'))).not.toBeInTheDocument();
    expect(screen.getByText('Advanced Analytics')).toBeInTheDocument();
  });

  it('does not show the empty state when a search matches an addon', () => {
    addonsData.data = [makeAddon({})];
    render(<Addons />);

    fireEvent.change(screen.getByPlaceholderText(i18next.t('common.management.searchAddons')), {
      target: { value: 'analytics' },
    });

    expect(screen.queryByText(i18next.t('addons.emptyMatchTitle'))).not.toBeInTheDocument();
  });
});
