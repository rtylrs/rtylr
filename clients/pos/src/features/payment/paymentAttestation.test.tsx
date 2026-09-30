/**
 * Two money defects live on this screen.
 *
 * 1. There is no terminal or payment-provider integration, so card/QR sales are
 *    recorded on the operator's word alone. The Charge button must therefore
 *    stay disabled until the operator explicitly attests, the screen must say
 *    the sale is operator-attested rather than verified, and an attestation
 *    must not survive a change of method or amount.
 * 2. A tip cannot be represented in the payment contract, so charging one books
 *    it as `change_due` server-side. The charge sent to the server must equal
 *    the tip-free order total the server will compute, never more.
 */
import React from 'react';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

const createOrder = vi.hoisted(() => vi.fn());
const processPayment = vi.hoisted(() => vi.fn());
const navigate = vi.hoisted(() => vi.fn());
const toastMock = vi.hoisted(() => ({
  success: vi.fn(),
  error: vi.fn(),
  warning: vi.fn(),
}));
const recordCashSale = vi.hoisted(() => vi.fn());
const addReceipt = vi.hoisted(() => vi.fn());
const enqueue = vi.hoisted(() => vi.fn());
const printReceipt = vi.hoisted(() => vi.fn());
const openDrawer = vi.hoisted(() => vi.fn());

const cart = vi.hoisted(() => ({
  items: [] as unknown[],
  subtotal: (): number => 0,
  taxTotal: (): number => 0,
  discountTotal: (): number => 0,
  total: (): number => 0,
  tipAmount: 0 as number,
  clearCart: vi.fn(),
  newOrder: vi.fn(),
  customerId: null as number | null,
  customerName: null as string | null,
  orderType: 'dine_in',
  tableId: null as number | null,
  notes: '',
  kitchenNotes: '',
  serverOrderId: null as number | null,
  setServerOrderId: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: (client: { name?: string }) =>
    client.name === 'OrderPaymentServiceClient'
      ? { processPayment }
      : { createOrder },
}));

vi.mock('react-router-dom', () => ({ useNavigate: () => navigate }));

vi.mock('sonner', () => ({ toast: toastMock }));

vi.mock('../../layouts/POSLayout', () => ({
  default: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));

vi.mock('motion/react', () => ({
  motion: new Proxy(
    {},
    {
      get: (_t, tag: string) =>
        ({ children, ...rest }: Record<string, unknown>) => {
          const Tag = tag as keyof React.JSX.IntrinsicElements;
          return React.createElement(Tag, rest, children as React.ReactNode);
        },
    },
  ),
}));

vi.mock('qrcode', () => ({ default: { toCanvas: vi.fn(() => Promise.resolve()) } }));

vi.mock('./paymentHardware', () => ({
  printReceiptAndOpenDrawer: printReceipt,
  openDrawer,
}));

vi.mock('../loyalty/loyaltyWorkflow', () => ({
  earnPointsAfterPayment: vi.fn(() => Promise.resolve(null)),
}));

vi.mock('../close-session/closeSessionWorkflow', () => ({
  resolveOpenSessionId: vi.fn(() => Promise.resolve(5)),
}));

vi.mock('../orders/orderCompatWorkflow', () => ({
  buildOrderCompatPayload: vi.fn(),
  buildResumeOrderPayload: vi.fn(),
  createOrderCompat: vi.fn(),
  requiresCompatOrder: () => false,
  persistResumedOrder: vi.fn(() => Promise.resolve()),
}));

vi.mock('../../shared/stores/cartStore', () => ({
  useCartStore: (selector: (s: typeof cart) => unknown) => selector(cart),
}));

vi.mock('../../shared/stores/settingsStore', () => ({
  useSettingsStore: (selector: (s: Record<string, unknown>) => unknown) =>
    selector({
      printer: { name: 'Default', copies: 1, cashDrawerKick: false, drawerPin: 2 },
      business: { name: 'Test Cafe' },
      features: { tables: false, cashDrawer: false, tips: true },
    }),
}));

vi.mock('../../shared/stores/offlineQueueStore', () => ({
  useOfflineQueueStore: (selector: (s: Record<string, unknown>) => unknown) =>
    selector({ enqueue }),
}));

vi.mock('../../shared/stores/terminalContextStore', () => ({
  useTerminalContextStore: (selector: (s: Record<string, unknown>) => unknown) =>
    selector({ locationId: 1 }),
}));

vi.mock('../../shared/stores/receiptsStore', () => ({
  useReceiptsStore: (selector: (s: Record<string, unknown>) => unknown) =>
    selector({ add: addReceipt }),
}));

vi.mock('../../shared/stores/cashDrawerStore', () => ({
  useCashDrawerStore: (selector: (s: Record<string, unknown>) => unknown) =>
    selector({ recordCashSale }),
}));

import Payment from './index';

const burger = {
  lineId: '1::',
  id: '1::',
  productId: 1,
  name: 'Burger',
  unitPrice: 20,
  price: 20,
  quantity: 1,
  modifiers: [],
  discountPct: 0,
  discountAmount: 0,
  taxRate: 0.08,
  status: 'pending' as const,
  notes: '',
};

/** $20.00 burger + 8% tax = $21.60 due. */
function seedCart({ storedTip = 0 }: { storedTip?: number } = {}) {
  cart.items = [burger];
  cart.subtotal = () => 20;
  cart.taxTotal = () => 1.6;
  cart.discountTotal = () => 0;
  cart.total = () => 21.6 + storedTip;
  cart.tipAmount = storedTip;
  cart.serverOrderId = null;
}

const chargeButton = () =>
  screen.getByRole('button', { name: /Confirm Payment ·/ }) as HTMLButtonElement;

const selectMethod = (label: string) => fireEvent.click(screen.getByText(label));

beforeEach(() => {
  vi.clearAllMocks();
  seedCart();
  createOrder.mockResolvedValue({ orderId: '500' });
  processPayment.mockResolvedValue({ message: 'Payment processed' });
  printReceipt.mockResolvedValue({ kind: 'ok' });
});

afterEach(() => cleanup());

describe('card/QR payment requires an explicit operator attestation', () => {
  it('keeps Charge disabled for a card tender until the operator attests', async () => {
    render(<Payment />);
    // Card is the default tender and used to be chargeable immediately.
    expect(chargeButton().disabled).toBe(true);

    fireEvent.click(screen.getByText('Tap to attest that you took this card payment yourself'));

    await waitFor(() => expect(chargeButton().disabled).toBe(false));
  });

  it('tells the cashier the sale is operator-attested, not provider-verified', () => {
    render(<Payment />);
    expect(
      screen.getByText(/No card terminal or payment provider is connected/),
    ).toBeTruthy();
    expect(screen.getByText(/Operator-attested\. No terminal or provider confirms/)).toBeTruthy();
  });

  it('voids the card attestation when the cashier switches tender and comes back', async () => {
    render(<Payment />);
    fireEvent.click(screen.getByText('Tap to attest that you took this card payment yourself'));
    await waitFor(() => expect(chargeButton().disabled).toBe(false));

    selectMethod('Cash');
    selectMethod('Card');

    await waitFor(() => expect(chargeButton().disabled).toBe(true));
  });

  it('voids the attestation when the order amount changes', async () => {
    render(<Payment />);
    fireEvent.click(screen.getByText('Tap to attest that you took this card payment yourself'));
    await waitFor(() => expect(chargeButton().disabled).toBe(false));

    // The order grows while the cashier is on the cash screen (e.g. an item is
    // rung in). The old attestation was for a different amount and must not hold.
    cart.taxTotal = () => 2.4;
    cart.total = () => 22.4;
    selectMethod('Cash');
    selectMethod('Card');

    await waitFor(() => expect(chargeButton().disabled).toBe(true));
  });

  it('keeps Charge disabled for QR until the operator attests, and never says "confirmed"', async () => {
    render(<Payment />);
    selectMethod('QR Pay');
    expect(chargeButton().disabled).toBe(true);
    expect(screen.queryByText('Confirm payment')).toBeNull();

    fireEvent.click(screen.getByText('I checked this payment elsewhere'));
    await waitFor(() => expect(chargeButton().disabled).toBe(false));
    expect(screen.getByText('Attested paid')).toBeTruthy();
  });

  it('sends the attested card amount to the server', async () => {
    render(<Payment />);
    fireEvent.click(screen.getByText('Tap to attest that you took this card payment yourself'));
    await waitFor(() => expect(chargeButton().disabled).toBe(false));

    fireEvent.click(chargeButton());

    await waitFor(() => expect(processPayment).toHaveBeenCalledTimes(1));
    expect(processPayment).toHaveBeenCalledWith(
      {
        orderId: '500',
        paymentMethod: 'card',
        amountPaid: 21.6,
      },
      // Required on this route: the server answers 428 without it, and dedups
      // the retry on it.
      { headers: { 'Idempotency-Key': expect.any(String) } },
    );
  });

  it('charges once when the button is double-tapped', async () => {
    render(<Payment />);
    fireEvent.click(screen.getByText('Tap to attest that you took this card payment yourself'));
    await waitFor(() => expect(chargeButton().disabled).toBe(false));

    // Both taps land before the disabled attribute is committed, and the label
    // has already flipped to "Processing…", so the same node is clicked twice.
    // `processing` state cannot stop that; the second tap must be refused
    // synchronously, because two checkouts are two orders and two charges.
    const button = chargeButton();
    fireEvent.click(button);
    fireEvent.click(button);

    await waitFor(() => expect(processPayment).toHaveBeenCalledTimes(1));
    expect(createOrder).toHaveBeenCalledTimes(1);
  });

  it('sends a different key for a second, separate sale', async () => {
    const { unmount } = render(<Payment />);
    fireEvent.click(screen.getByText('Tap to attest that you took this card payment yourself'));
    await waitFor(() => expect(chargeButton().disabled).toBe(false));
    fireEvent.click(chargeButton());
    await waitFor(() => expect(processPayment).toHaveBeenCalledTimes(1));
    unmount();

    seedCart();
    render(<Payment />);
    fireEvent.click(screen.getByText('Tap to attest that you took this card payment yourself'));
    await waitFor(() => expect(chargeButton().disabled).toBe(false));
    fireEvent.click(chargeButton());
    await waitFor(() => expect(processPayment).toHaveBeenCalledTimes(2));

    const keys = processPayment.mock.calls.map((call) => (call[1] as { headers: Record<string, string> }).headers['Idempotency-Key']);
    expect(keys[0]).not.toBe(keys[1]);
  });

  it('never reaches the API for an unattested card tender', async () => {
    render(<Payment />);
    expect(chargeButton().disabled).toBe(true);

    // A double-tap burst that lands before the disabled attribute is committed
    // must still not create or charge an order.
    fireEvent.click(chargeButton());
    fireEvent.click(chargeButton());

    await Promise.resolve();
    expect(createOrder).not.toHaveBeenCalled();
    expect(processPayment).not.toHaveBeenCalled();
  });
});

describe('a tip can never be charged as change', () => {
  it('offers no tip control on the checkout screen', () => {
    render(<Payment />);
    expect(screen.queryByText('Tip')).toBeNull();
  });

  it('charges the tip-free total even when the cart carries a stored tip', async () => {
    // A resumed held order restores its tip into the cart, and `total()`
    // includes it. Charging `total()` would inflate amount_paid by the tip and
    // the server would store the difference as change_due.
    seedCart({ storedTip: 4 });
    render(<Payment />);
    fireEvent.click(screen.getByText('Tap to attest that you took this card payment yourself'));
    await waitFor(() => expect(chargeButton().disabled).toBe(false));

    fireEvent.click(chargeButton());

    await waitFor(() => expect(processPayment).toHaveBeenCalledTimes(1));
    expect(processPayment).toHaveBeenCalledWith(
      {
        orderId: '500',
        paymentMethod: 'card',
        amountPaid: 21.6,
      },
      { headers: { 'Idempotency-Key': expect.any(String) } },
    );
  });
});
