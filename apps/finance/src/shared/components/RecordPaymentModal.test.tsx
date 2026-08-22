// @vitest-environment happy-dom
import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import RecordPaymentModal from './RecordPaymentModal';

const recordPaymentMock = vi.fn();

vi.mock('../../features/invoices/useInvoices', () => ({
  useRecordPayment: () => ({ mutate: recordPaymentMock, isPending: false }),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('./Modal', () => ({
  default: ({ children, footer }: { children?: React.ReactNode; footer?: React.ReactNode }) => (
    <div>
      {children}
      {footer}
    </div>
  ),
}));

vi.mock('./CurrencyInput', () => ({
  default: ({ value, onChange }: { value: number; onChange: (v: number) => void }) => (
    <input
      aria-label="payments.amount"
      value={String(value)}
      onChange={(e) => {
        const v = parseFloat(e.target.value);
        if (!Number.isNaN(v)) onChange(v);
      }}
    />
  ),
}));

describe('RecordPaymentModal', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  const setup = (props?: Partial<Parameters<typeof RecordPaymentModal>[0]>) =>
    render(
      <RecordPaymentModal
        open
        onClose={vi.fn()}
        invoiceId="7"
        amount={100}
        currency="USD"
        {...props}
      />
    );

  it('offers the canonical lowercase payment methods', () => {
    setup();
    const select = screen.getByRole('combobox', { name: 'payments.method' }) as HTMLSelectElement;
    expect(Array.from(select.options).map((o) => o.value)).toEqual([
      'cash',
      'card',
      'bank_transfer',
    ]);
  });

  it('disables submit when the invoice currency is unknown', () => {
    setup({ currency: null });
    expect(screen.getByText('payments.currencyUnknownHint')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'payments.record' })).toBeDisabled();
  });

  it('blocks amounts exceeding the outstanding balance and shows why', () => {
    setup({ amount: 100 });
    fireEvent.change(screen.getByLabelText('payments.amount'), { target: { value: '150' } });
    expect(screen.getByRole('alert')).toHaveTextContent('payments.exceedsBalanceHint');
    expect(screen.getByRole('button', { name: 'payments.record' })).toBeDisabled();
  });

  it('submits a valid payment in the invoice currency with the selected method', () => {
    setup({ amount: 100 });
    fireEvent.change(screen.getByLabelText(/payments.reference/), {
      target: { value: 'TXN-1' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'payments.record' }));
    expect(recordPaymentMock).toHaveBeenCalledWith(
      expect.objectContaining({
        id: '7',
        amount: 100,
        currency: 'USD',
        method: 'cash',
        reference: 'TXN-1',
      }),
      expect.anything()
    );
  });
});
