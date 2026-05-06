import { formatCurrency } from '../../shared/utils/format';
import type {
  QueuedCashMovement,
  QueuedCashMovementAction,
} from '../../shared/stores/cashMovementQueueStore';
import type { CashMovementPayload } from '../../shared/services/posService';

export type CashDrawerAction = 'drop' | 'paid_out' | 'paid_in' | 'float';

interface CashDrawerActions {
  openDrawer: (amount: number) => void;
  cashDrop: (amount: number, reason?: string) => void;
  paidOut: (amount: number, reason: string) => void;
  paidIn: (amount: number, reason: string) => void;
}

interface RecordCashDrawerActionParams {
  action: CashDrawerAction;
  amountInput: string;
  reason: string;
  drawer: CashDrawerActions;
}

interface RecordCashDrawerActionResult {
  amount: number;
  message: string;
}

interface ProcessCashDrawerActionParams extends RecordCashDrawerActionParams {
  syncMovement: (action: QueuedCashMovementAction, payload: CashMovementPayload) => Promise<void>;
  enqueueMovement: (
    movement: Pick<QueuedCashMovement, 'action' | 'payload'>
  ) => QueuedCashMovement | void;
}

interface ProcessCashDrawerActionResult extends RecordCashDrawerActionResult {
  queued: boolean;
}

export function recordCashDrawerAction({
  action,
  amountInput,
  reason,
  drawer,
}: RecordCashDrawerActionParams): RecordCashDrawerActionResult {
  const amount = Number(amountInput);
  if (!Number.isFinite(amount) || amount <= 0) {
    throw new Error('Enter a valid amount greater than zero.');
  }

  const trimmedReason = reason.trim();
  if (action !== 'float' && !trimmedReason) {
    throw new Error('Enter a reason for this cash movement.');
  }

  if (action === 'float') {
    drawer.openDrawer(amount);
    return { amount, message: `Opening float set to ${formatCurrency(amount)}` };
  }

  if (action === 'drop') {
    drawer.cashDrop(amount, trimmedReason);
    return { amount, message: 'Cash drop recorded' };
  }

  if (action === 'paid_out') {
    drawer.paidOut(amount, trimmedReason);
    return { amount, message: 'Paid out' };
  }

  drawer.paidIn(amount, trimmedReason);
  return { amount, message: 'Paid in' };
}

export async function processCashDrawerAction({
  action,
  amountInput,
  reason,
  drawer,
  syncMovement,
  enqueueMovement,
}: ProcessCashDrawerActionParams): Promise<ProcessCashDrawerActionResult> {
  const result = recordCashDrawerAction({ action, amountInput, reason, drawer });
  if (action === 'float') return { ...result, queued: false };

  const payload = { amount: result.amount, reason: reason.trim() };
  try {
    await syncMovement(action, payload);
    return { ...result, queued: false };
  } catch {
    enqueueMovement({ action, payload });
    return { ...result, queued: true };
  }
}
