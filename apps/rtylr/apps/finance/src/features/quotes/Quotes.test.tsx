// @vitest-environment happy-dom
import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi, beforeAll } from 'vitest';
import Quotes from './Quotes';
import { initTestI18n } from '../../shared/testing/i18n';

const quote = {
  id: 'Q-100',
  client: 'Acme Corp',
  amount: 1250,
  status: 'Sent',
  issuedDate: '2026-06-24T08:50:45Z',
  expiresAt: '2026-07-24',
};

const { useQuotesMock } = vi.hoisted(() => ({ useQuotesMock: vi.fn() }));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('motion/react', () => ({
  motion: {
    div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div>,
  },
}));

// i18next is deliberately not mocked. An identity `t` makes an assertion and
// the component agree on a key, so neither notices the key never resolves
// (rtylr#1269). The real locale files mean the copy below also proves the lookup.

vi.mock('sonner', () => ({ toast: { error: vi.fn(), success: vi.fn() } }));

vi.mock('./useQuotes', () => ({
  useQuotes: useQuotesMock,
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

vi.mock('../../shared/components/Skeleton', () => ({ SkeletonRow: () => null }));
vi.mock('../../shared/components/Modal', () => ({
  default: ({
    open,
    title,
    children,
  }: {
    open?: boolean;
    title?: string;
    children?: React.ReactNode;
  }) =>
    open ? (
      <div role="dialog">
        {title}
        {children}
      </div>
    ) : null,
}));
vi.mock('../../shared/components/CurrencyInput', () => ({ default: () => null }));

beforeAll(async () => {
  await initTestI18n('en');
});

describe('Quotes table', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQuotesMock.mockReturnValue({ data: { quotes: [quote], hasMore: false }, isLoading: false });
  });

  it('formats the Issued timestamp as a readable date', () => {
    render(<Quotes />);

    expect(screen.getByRole('cell', { name: 'Jun 24, 2026' })).toBeInTheDocument();
    expect(screen.queryByText('2026-06-24T08:50:45Z')).not.toBeInTheDocument();
  });
});

describe('Quotes empty state', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQuotesMock.mockReturnValue({ data: { quotes: [], hasMore: false }, isLoading: false });
  });

  it('renders the shared EmptyState CTA and opens the new quote modal on click', () => {
    render(<Quotes />);

    expect(screen.queryByRole('table')).not.toBeInTheDocument();
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();

    const newQuoteButtons = screen.getAllByRole('button', { name: 'New Quote' });
    expect(newQuoteButtons).toHaveLength(2);
    fireEvent.click(newQuoteButtons[1]);

    expect(screen.getByRole('dialog')).toHaveTextContent('New Quote');
  });

  it('does not offer an Issue Date the create contract cannot store', () => {
    render(<Quotes />);

    fireEvent.click(screen.getAllByRole('button', { name: 'New Quote' })[1]);

    const dialog = screen.getByRole('dialog');
    expect(dialog).toHaveTextContent('Expires');
    expect(dialog).not.toHaveTextContent('Issue Date');
  });
});
