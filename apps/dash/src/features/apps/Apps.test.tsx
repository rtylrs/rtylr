import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { Grid3x3 } from 'lucide-react';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('motion/react', () => ({
  motion: {
    a: (props: React.ComponentProps<'a'>) => <a {...props} />,
    div: (props: React.ComponentProps<'div'>) => <div {...props} />,
  },
}));

type App = {
  key: string;
  label: string;
  description: string;
  category: string;
  icon: typeof Grid3x3;
  bg: string;
  color: string;
  href: string;
};

const makeApp = (overrides: Partial<App>): App => ({
  key: 'erp',
  label: 'ERP',
  description: 'Inventory',
  category: 'operations',
  icon: Grid3x3,
  bg: '#000',
  color: '#fff',
  href: '/erp',
  ...overrides,
});

const appsData = vi.hoisted(() => ({
  data: [] as App[],
  isLoading: false,
}));

vi.mock('../../shared/hooks/queries/useApps', () => ({
  useApps: () => appsData,
}));

vi.mock('../../shared/stores/userStore', () => ({
  useUserStore: () => ({ user: { role: 'owner', apps: [] } }),
}));

import Apps from './Apps';

beforeEach(() => {
  vi.clearAllMocks();
  appsData.data = [];
  appsData.isLoading = false;
});

describe('Apps grid empty state', () => {
  it('shows the illustration, title and a clear-filters call-to-action when no apps match', () => {
    render(<Apps />);

    expect(screen.getByText('apps.emptyTitle')).toBeInTheDocument();
    expect(screen.getByText('apps.emptyDescription')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'common.clearFilters' })).toBeInTheDocument();
  });

  it('clears the search query and category when the call-to-action is clicked', () => {
    appsData.data = [makeApp({})];
    render(<Apps />);

    fireEvent.change(screen.getByPlaceholderText('Search apps...'), {
      target: { value: 'nonexistent app' },
    });
    expect(screen.getByText('apps.emptyTitle')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'common.clearFilters' }));

    expect(screen.queryByText('apps.emptyTitle')).not.toBeInTheDocument();
    expect(screen.getByText('ERP')).toBeInTheDocument();
  });

  it('does not show the empty state when apps match', () => {
    appsData.data = [makeApp({})];
    render(<Apps />);

    expect(screen.queryByText('apps.emptyTitle')).not.toBeInTheDocument();
  });
});
