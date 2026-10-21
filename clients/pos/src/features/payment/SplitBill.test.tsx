/**
 * The "By amount" split summed each field with parseFloat, which stops at the
 * first character it cannot parse. In a French terminal "12,50" summed as 12,
 * so the screen showed a sum 100x below what the operator had typed and would
 * have split a ticket that does not add up. The mode cannot reach the API
 * today - confirm() returns early unless the mode is 'seat' - so this is a
 * landmine, not yet a live charge, and it is fixed here so it is not one when
 * the mode ships.
 */
import React from 'react';
import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

const splitOrder = vi.hoisted(() => vi.fn());
const navigate = vi.hoisted(() => vi.fn());
let currentLanguage = 'en';

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => ({ splitOrder }),
}));

vi.mock('react-router-dom', () => ({
  useNavigate: () => navigate,
  useParams: () => ({ orderId: '42' }),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    i18n: { resolvedLanguage: currentLanguage, language: currentLanguage },
  }),
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn(), success: vi.fn(), warning: vi.fn() },
}));

vi.mock('../../layouts/ScreenLayout', () => ({
  default: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));

vi.mock('../../shared/stores/cartStore', () => ({
  useCartStore: (selector: (state: unknown) => unknown) =>
    selector({ items: [], total: () => 25, serverOrderId: 42 }),
}));

import SplitBillScreen from './SplitBill';

/** Opens the amount mode and returns its inputs, left-to-right. */
function openAmountMode(): HTMLInputElement[] {
  render(<SplitBillScreen />);
  fireEvent.click(screen.getByText('By amount'));
  return screen.getAllByRole('spinbutton') as HTMLInputElement[];
}

describe('SplitBill amount mode', () => {
  beforeEach(() => {
    currentLanguage = 'en';
  });
  afterEach(() => {
    cleanup();
    vi.clearAllMocks();
  });

  it('sums a French amount at its true value, not at the truncated one', () => {
    currentLanguage = 'fr';
    const [first, second] = openAmountMode();

    fireEvent.change(first, { target: { value: '12,50' } });
    fireEvent.change(second, { target: { value: '12,50' } });

    // 12 + 12 = 24 read by parseFloat, against a $25.00 ticket.
    expect(Number.parseFloat('12,50')).toBe(12);
    expect(screen.getByText(/Sum:/).textContent).toContain('25.00');
    expect(screen.getByText(/Sum:/).textContent).not.toContain('24.00');
  });

  it('sums an English amount at its true value', () => {
    const [first, second] = openAmountMode();

    fireEvent.change(first, { target: { value: '12.50' } });
    fireEvent.change(second, { target: { value: '12.50' } });

    expect(screen.getByText(/Sum:/).textContent).toContain('25.00');
  });

  it('flags a genuine mismatch instead of hiding it', () => {
    const [first, second] = openAmountMode();

    fireEvent.change(first, { target: { value: '10' } });
    fireEvent.change(second, { target: { value: '10' } });

    expect(screen.getByText(/must match the total/)).toBeTruthy();
  });

  it('keeps the amount the operator typed on screen', () => {
    currentLanguage = 'fr';
    const [first] = openAmountMode();

    fireEvent.change(first, { target: { value: '12,50' } });

    expect(first.value).toBe('12,50');
  });

  it('reads a grouped French amount with its thousands separator', () => {
    currentLanguage = 'fr';
    const [first, second] = openAmountMode();

    fireEvent.change(first, { target: { value: '1.234,50' } });
    fireEvent.change(second, { target: { value: '0,00' } });

    expect(screen.getByText(/must match the total/)).toBeTruthy();
    expect(screen.getByText(/Sum:/).textContent).toContain('1,234.50');
  });

  it('still refuses to confirm an amount split that is not finished', () => {
    openAmountMode();

    const confirm = screen.getByText('Confirm Split') as HTMLButtonElement;
    expect(confirm.disabled).toBe(true);
    expect(splitOrder).not.toHaveBeenCalled();
  });
});
