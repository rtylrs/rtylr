/**
 * A double-submitted refund pays the customer twice, so the confirm dialog
 * must be idempotent: disabled while in flight, guarded in the handler, and
 * re-armed after a failure so the cashier can retry.
 */
import React from 'react';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import RefundScreen from './index';

const refundClient = vi.hoisted(() => ({ processRefund: vi.fn() }));
const orderClient = vi.hoisted(() => ({ getOrderDetail: vi.fn() }));
const recordCashRefund = vi.hoisted(() => vi.fn());
const navigate = vi.hoisted(() => vi.fn());
const toastMock = vi.hoisted(() => ({ success: vi.fn(), error: vi.fn() }));

vi.mock('@voxire/api', () => ({
  createOneKitClient: (client: { name?: string }) => {
    if (client.name === 'OrderRefundServiceClient') return refundClient;
    if (client.name === 'OrderServiceClient') return orderClient;
    return {};
  },
}));

vi.mock('react-router-dom', () => ({
  useParams: () => ({ orderId: '7' }),
  useNavigate: () => navigate,
}));

vi.mock('sonner', () => ({ toast: toastMock }));

vi.mock('../../layouts/ScreenLayout', () => ({
  default: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));

vi.mock('../../shared/components/ManagerRequired', () => ({ default: () => <div /> }));

vi.mock('../../shared/hooks/useCapabilities', () => ({
  useCapabilities: () => ({ canRefund: true }),
}));

vi.mock('../../shared/stores/receiptsStore', () => ({
  useReceiptsStore: (selector: (s: { byOrder: () => undefined }) => unknown) =>
    selector({ byOrder: () => undefined }),
}));

vi.mock('../../shared/stores/cashDrawerStore', () => ({
  useCashDrawerStore: (selector: (s: { recordCashRefund: () => void }) => unknown) =>
    selector({ recordCashRefund }),
}));

async function openConfirmDialog() {
  render(<RefundScreen />);
  fireEvent.click(await screen.findByText('Issue Full Refund'));
  return screen.findByRole('button', { name: 'Refund' });
}

describe('RefundScreen confirm dialog', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    orderClient.getOrderDetail.mockResolvedValue({ order: { total: 50 } });
    refundClient.processRefund.mockResolvedValue(undefined);
  });

  afterEach(() => {
    cleanup();
  });

  it('sends one refund when confirm is double-clicked', async () => {
    // Never resolves during the test, so the second click lands mid-flight.
    let release: (() => void) | undefined;
    refundClient.processRefund.mockImplementation(
      () =>
        new Promise<void>((resolve) => {
          release = resolve;
        })
    );

    const confirm = await openConfirmDialog();
    fireEvent.click(confirm);
    fireEvent.click(confirm);
    fireEvent.click(confirm);

    expect(refundClient.processRefund).toHaveBeenCalledTimes(1);

    release?.();
    await waitFor(() => expect(navigate).toHaveBeenCalledWith('/receipts'));
  });

  it('disables both dialog buttons while the refund is in flight', async () => {
    let release: (() => void) | undefined;
    refundClient.processRefund.mockImplementation(
      () =>
        new Promise<void>((resolve) => {
          release = resolve;
        })
    );

    const confirm = await openConfirmDialog();
    fireEvent.click(confirm);

    await waitFor(() => expect((confirm as HTMLButtonElement).disabled).toBe(true));
    expect((screen.getByRole('button', { name: 'Cancel' }) as HTMLButtonElement).disabled).toBe(true);

    release?.();
    await waitFor(() => expect(navigate).toHaveBeenCalledWith('/receipts'));
  });

  it('does not dismiss the dialog on an outside click while in flight', async () => {
    let release: (() => void) | undefined;
    refundClient.processRefund.mockImplementation(
      () =>
        new Promise<void>((resolve) => {
          release = resolve;
        })
    );

    const confirm = await openConfirmDialog();
    fireEvent.click(confirm);
    await waitFor(() => expect((confirm as HTMLButtonElement).disabled).toBe(true));

    fireEvent.click(screen.getByRole('dialog'));
    expect(screen.getByRole('dialog')).toBeTruthy();

    fireEvent.keyDown(document, { key: 'Escape' });
    expect(screen.getByRole('dialog')).toBeTruthy();

    release?.();
    await waitFor(() => expect(navigate).toHaveBeenCalledWith('/receipts'));
  });

  it('re-enables confirm after a failure so the cashier can retry', async () => {
    refundClient.processRefund
      .mockRejectedValueOnce(new Error('Network error'))
      .mockResolvedValueOnce(undefined);

    const confirm = await openConfirmDialog();
    fireEvent.click(confirm);

    await waitFor(() => expect(toastMock.error).toHaveBeenCalledWith('Network error'));
    // Dialog stays open and the button is live again.
    expect((confirm as HTMLButtonElement).disabled).toBe(false);
    expect(refundClient.processRefund).toHaveBeenCalledTimes(1);

    fireEvent.click(confirm);

    await waitFor(() => expect(refundClient.processRefund).toHaveBeenCalledTimes(2));
    expect(navigate).toHaveBeenCalledWith('/receipts');
  });
});
