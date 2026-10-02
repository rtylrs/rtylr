// @vitest-environment happy-dom
import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import { initTestI18n } from '../../shared/testing/i18n';

const { matchReconcileMock, toastMock } = vi.hoisted(() => ({
  matchReconcileMock: vi.fn(),
  toastMock: { success: vi.fn(), error: vi.fn() },
}));

const transactions = [
  {
    id: 'tx-1',
    matched: false,
    description: 'Acme payment',
    amount: 500,
    date: '2026-08-01',
    category: 'Sales',
  },
  {
    id: 'tx-2',
    matched: false,
    description: 'Globex payment',
    amount: 1200,
    date: '2026-08-02',
    category: 'Sales',
  },
  {
    id: 'tx-3',
    matched: false,
    description: 'Initech payment',
    amount: 750,
    date: '2026-08-03',
    category: 'Sales',
  },
];

const invoices = [
  { id: 'inv-1', invoiceNumber: 'INV-1', clientName: 'Acme', total: 500 },
  { id: 'inv-2', invoiceNumber: 'INV-2', clientName: 'Globex', total: 1200 },
  { id: 'inv-3', invoiceNumber: 'INV-3', clientName: 'Initech', total: 750 },
];

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));

vi.mock('motion/react', () => ({
  motion: { div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div> },
}));

vi.mock('sonner', () => ({ toast: toastMock }));

// Keep the real unwrap/error-normalisation shape out of the way; the calls
// under test are the hook's sequencing of successful and failed matches.
vi.mock('@voxire/api', () => ({
  getOneKitErrorMessage: (_error: unknown, fallback: string) => fallback,
  unwrap: (result: Promise<unknown>) => result,
}));

vi.mock('../../shared/hooks/useBankAccounts', () => ({
  useBankTransactions: () => ({ data: transactions }),
}));

vi.mock('../../shared/hooks/queries/useInvoices', () => ({
  useListInvoices: () => ({ data: { invoices } }),
}));

vi.mock('../../shared/services/financeApi', () => ({
  financeApi: { matchReconcile: matchReconcileMock },
}));

vi.mock('../../shared/components/PageHeader', () => ({
  default: ({ actions }: { actions?: React.ReactNode }) => <div>{actions}</div>,
}));

vi.mock('../../shared/components/MoneyDisplay', () => ({
  default: ({ amount }: { amount: number }) => <span>{amount}</span>,
}));

import Reconcile from './Reconcile';

// i18next is deliberately not mocked so the assertions also prove the locale
// keys resolve (see Quotes.test.tsx and rtylr#1269).
beforeAll(async () => {
  await initTestI18n('en');
});

const clickAutoMatch = (): void => {
  fireEvent.click(screen.getByRole('button', { name: /Auto-match all/ }));
};

describe('Reconcile auto-match', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('keeps matches that already landed when a later reconcile call fails', async () => {
    let failGlobexOnce = true;
    matchReconcileMock.mockImplementation((transactionId: string) => {
      if (transactionId === 'tx-2' && failGlobexOnce) {
        failGlobexOnce = false;
        return Promise.reject(new Error('500'));
      }
      return Promise.resolve(undefined);
    });

    render(<Reconcile />);
    clickAutoMatch();

    // tx-1 was accepted by the server before tx-2 failed, so it must leave
    // the unmatched list instead of being silently hidden until reload.
    await waitFor(() => expect(screen.queryByText('Acme payment')).not.toBeInTheDocument());
    expect(screen.getByText('Globex payment')).toBeInTheDocument();
    expect(screen.getByText('Initech payment')).toBeInTheDocument();
    expect(matchReconcileMock).toHaveBeenNthCalledWith(1, 'tx-1', 'inv-1');
    expect(matchReconcileMock).toHaveBeenNthCalledWith(2, 'tx-2', 'inv-2');
    expect(matchReconcileMock).toHaveBeenCalledTimes(2);
    expect(toastMock.error).toHaveBeenCalledWith('Auto-match failed');

    // Retrying must not re-send the match the server already persisted.
    clickAutoMatch();
    await waitFor(() =>
      expect(screen.getByText('All transactions reconciled.')).toBeInTheDocument()
    );
    expect(matchReconcileMock).toHaveBeenCalledTimes(4);
    expect(matchReconcileMock).toHaveBeenNthCalledWith(3, 'tx-2', 'inv-2');
    expect(matchReconcileMock).toHaveBeenNthCalledWith(4, 'tx-3', 'inv-3');
  });
});
