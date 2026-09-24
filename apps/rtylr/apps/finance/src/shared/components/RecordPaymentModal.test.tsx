// @vitest-environment happy-dom
import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import RecordPaymentModal from './RecordPaymentModal';
import { initTestI18n } from '../testing/i18n';

const mocks = vi.hoisted(() => ({
  mutate: vi.fn(),
  toastError: vi.fn(),
  toastSuccess: vi.fn(),
}));

vi.mock('../../features/invoices/useInvoices', () => ({
  useRecordPayment: () => ({ mutate: mocks.mutate, isPending: false }),
}));

vi.mock('sonner', () => ({
  toast: { error: mocks.toastError, success: mocks.toastSuccess },
}));

vi.mock('@voxire/api', () => ({
  getOneKitErrorMessage: (_error: unknown, fallback: string) => fallback,
}));

vi.mock('./Modal', () => ({
  default: ({ children, footer }: { children?: React.ReactNode; footer?: React.ReactNode }) => (
    <div>
      {children}
      {footer}
    </div>
  ),
}));

beforeAll(async () => {
  await initTestI18n('en');
});

beforeEach(() => {
  vi.clearAllMocks();
  mocks.mutate.mockReset();
});

describe('RecordPaymentModal', () => {
  it('converts the displayed major-unit amount back to the API major-unit amount', () => {
    render(
      <RecordPaymentModal open onClose={vi.fn()} invoiceId="INV-1" amount={500} currency="USD" />
    );

    const amountInput = screen.getByDisplayValue('500');
    fireEvent.change(amountInput, { target: { value: '250' } });
    fireEvent.change(screen.getByPlaceholderText('Transaction ID or receipt number'), {
      target: { value: 'TXN-1' },
    });
    fireEvent.click(screen.getByRole('button', { name: /Record Payment/ }));

    expect(mocks.mutate).toHaveBeenCalledWith(
      expect.objectContaining({ id: 'INV-1', amount: 250, currency: 'USD' }),
      expect.any(Object)
    );
  });

  it('rejects a payment larger than the remaining balance', () => {
    render(
      <RecordPaymentModal open onClose={vi.fn()} invoiceId="INV-1" amount={500} currency="USD" />
    );

    fireEvent.change(screen.getByDisplayValue('500'), { target: { value: '600' } });
    fireEvent.change(screen.getByPlaceholderText('Transaction ID or receipt number'), {
      target: { value: 'TXN-1' },
    });
    fireEvent.click(screen.getByRole('button', { name: /Record Payment/ }));

    expect(mocks.toastError).toHaveBeenCalledWith('Payment cannot exceed the remaining amount due');
    expect(mocks.mutate).not.toHaveBeenCalled();
  });
});
