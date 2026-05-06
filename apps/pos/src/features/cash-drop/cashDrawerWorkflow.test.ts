import { describe, expect, it, vi } from 'vitest';
import { processCashDrawerAction, recordCashDrawerAction } from './cashDrawerWorkflow';

const actions = () => ({
  openDrawer: vi.fn(),
  cashDrop: vi.fn(),
  paidOut: vi.fn(),
  paidIn: vi.fn(),
});

describe('recordCashDrawerAction', () => {
  it('opens the drawer for a valid float', () => {
    const drawer = actions();

    const result = recordCashDrawerAction({ action: 'float', amountInput: '125.50', reason: '', drawer });

    expect(result).toEqual({ amount: 125.5, message: 'Opening float set to $125.50' });
    expect(drawer.openDrawer).toHaveBeenCalledWith(125.5);
  });

  it('rejects non-finite amounts before mutating drawer state', () => {
    const drawer = actions();

    expect(() =>
      recordCashDrawerAction({ action: 'paid_in', amountInput: 'abc', reason: 'Correction', drawer })
    ).toThrow('Enter a valid amount greater than zero.');
    expect(drawer.paidIn).not.toHaveBeenCalled();
  });

  it('requires a reason for manual cash movements', () => {
    const drawer = actions();

    expect(() =>
      recordCashDrawerAction({ action: 'drop', amountInput: '20', reason: '   ', drawer })
    ).toThrow('Enter a reason for this cash movement.');
    expect(drawer.cashDrop).not.toHaveBeenCalled();
  });

  it('records a cash drop with trimmed reason', () => {
    const drawer = actions();

    const result = recordCashDrawerAction({
      action: 'drop',
      amountInput: '20',
      reason: ' Safe deposit ',
      drawer,
    });

    expect(result).toEqual({ amount: 20, message: 'Cash drop recorded' });
    expect(drawer.cashDrop).toHaveBeenCalledWith(20, 'Safe deposit');
  });
});

describe('processCashDrawerAction', () => {
  it('syncs manual cash movements after recording them locally', async () => {
    const drawer = actions();
    const syncMovement = vi.fn(async () => undefined);
    const enqueueMovement = vi.fn();

    const result = await processCashDrawerAction({
      action: 'paid_out',
      amountInput: '15',
      reason: 'Petty cash',
      drawer,
      syncMovement,
      enqueueMovement,
    });

    expect(result).toEqual({ amount: 15, message: 'Paid out', queued: false });
    expect(drawer.paidOut).toHaveBeenCalledWith(15, 'Petty cash');
    expect(syncMovement).toHaveBeenCalledWith('paid_out', { amount: 15, reason: 'Petty cash' });
    expect(enqueueMovement).not.toHaveBeenCalled();
  });

  it('enqueues manual cash movements when backend sync fails', async () => {
    const drawer = actions();
    const enqueueMovement = vi.fn();

    const result = await processCashDrawerAction({
      action: 'drop',
      amountInput: '20',
      reason: 'Safe deposit',
      drawer,
      syncMovement: vi.fn(async () => {
        throw new Error('offline');
      }),
      enqueueMovement,
    });

    expect(result).toEqual({ amount: 20, message: 'Cash drop recorded', queued: true });
    expect(drawer.cashDrop).toHaveBeenCalledWith(20, 'Safe deposit');
    expect(enqueueMovement).toHaveBeenCalledWith({
      action: 'drop',
      payload: { amount: 20, reason: 'Safe deposit' },
    });
  });

  it('does not sync opening float as a cash movement', async () => {
    const drawer = actions();
    const syncMovement = vi.fn();

    await processCashDrawerAction({
      action: 'float',
      amountInput: '100',
      reason: '',
      drawer,
      syncMovement,
      enqueueMovement: vi.fn(),
    });

    expect(syncMovement).not.toHaveBeenCalled();
  });
});
