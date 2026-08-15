// @vitest-environment happy-dom
import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import Quotes from './Quotes';

const quote = {
  id: 'Q-100',
  client: 'Acme Corp',
  amount: 1250,
  status: 'Sent',
  issuedDate: '2026-06-24T08:50:45Z',
  expiresAt: '2026-07-24',
};

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('motion/react', () => ({
  motion: {
    div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div>,
  },
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('sonner', () => ({ toast: { error: vi.fn(), success: vi.fn() } }));

vi.mock('./useQuotes', () => ({
  useQuotes: () => ({ data: { quotes: [quote], hasMore: false }, isLoading: false }),
  useCreateQuote: () => ({ mutate: vi.fn(), isPending: false }),
  useConvertQuote: () => ({ mutate: vi.fn(), isPending: false }),
  useUpdateQuoteStatus: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('../../shared/components/PageHeader', () => ({
  default: ({ actions }: { actions?: React.ReactNode }) => <div>{actions}</div>,
}));

vi.mock('../../shared/components/MoneyDisplay', () => ({
  default: ({ amount }: { amount: number }) => <span>{amount}</span>,
}));

vi.mock('../../shared/components/EmptyState', () => ({ default: () => null }));
vi.mock('../../shared/components/Skeleton', () => ({ SkeletonRow: () => null }));
vi.mock('../../shared/components/Modal', () => ({ default: () => null }));
vi.mock('../../shared/components/CurrencyInput', () => ({ default: () => null }));

describe('Quotes table', () => {
  it('formats the Issued timestamp as a readable date', () => {
    render(<Quotes />);

    expect(screen.getByRole('cell', { name: 'Jun 24, 2026' })).toBeInTheDocument();
    expect(screen.queryByText('2026-06-24T08:50:45Z')).not.toBeInTheDocument();
  });
});
