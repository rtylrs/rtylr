import { formatCurrency } from "../../shared/utils/format";
import { isRetryableTransportError } from "../../shared/utils/errorClassify";
import type {
  CashMovementPayload,
  QueuedCashMovement,
  QueuedCashMovementAction,
} from "../../shared/stores/cashMovementQueueStore";
import { NoOpenRegisterSessionError } from "../close-session/closeSessionWorkflow";

export type CashDrawerAction = "drop" | "paid_out" | "paid_in" | "float";

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

interface CashMovementContext {
  branchId: number | null;
  terminalId?: number | null;
  employeeId: number | null;
  resolveSessionId: () => Promise<number>;
}

function requireMovementContext(context: CashMovementContext): {
  branchId: number;
  employeeId: number;
} {
  if (!context.branchId || !context.employeeId) {
    throw new Error(
      "Select a terminal and sign in before recording cash movements.",
    );
  }
  return { branchId: context.branchId, employeeId: context.employeeId };
}

interface ProcessCashDrawerActionParams extends RecordCashDrawerActionParams {
  context: CashMovementContext;
  syncMovement: (
    action: QueuedCashMovementAction,
    payload: CashMovementPayload,
  ) => Promise<void>;
  enqueueMovement: (
    movement: Pick<QueuedCashMovement, "action" | "payload">,
  ) => QueuedCashMovement | void;
}

interface ProcessCashDrawerActionResult extends RecordCashDrawerActionResult {
  queued: boolean;
}

interface PreparedCashDrawerAction extends RecordCashDrawerActionResult {
  reason: string;
}

function prepareCashDrawerAction({
  action,
  amountInput,
  reason,
}: Omit<RecordCashDrawerActionParams, "drawer">): PreparedCashDrawerAction {
  const amount = Number(amountInput);
  if (!Number.isFinite(amount) || amount <= 0) {
    throw new Error("Enter a valid amount greater than zero.");
  }

  const trimmedReason = reason.trim();
  if (action !== "float" && !trimmedReason) {
    throw new Error("Enter a reason for this cash movement.");
  }

  let message: string;
  if (action === "float") {
    message = `Opening float set to ${formatCurrency(amount)}`;
  } else if (action === "drop") {
    message = "Cash drop recorded";
  } else if (action === "paid_out") {
    message = "Paid out";
  } else {
    message = "Paid in";
  }
  return { amount, reason: trimmedReason, message };
}

function applyCashDrawerAction(
  action: CashDrawerAction,
  amount: number,
  reason: string,
  drawer: CashDrawerActions,
): void {
  if (action === "float") {
    drawer.openDrawer(amount);
  } else if (action === "drop") {
    drawer.cashDrop(amount, reason);
  } else if (action === "paid_out") {
    drawer.paidOut(amount, reason);
  } else {
    drawer.paidIn(amount, reason);
  }
}

export function recordCashDrawerAction({
  action,
  amountInput,
  reason,
  drawer,
}: RecordCashDrawerActionParams): RecordCashDrawerActionResult {
  const prepared = prepareCashDrawerAction({ action, amountInput, reason });
  applyCashDrawerAction(action, prepared.amount, prepared.reason, drawer);
  return { amount: prepared.amount, message: prepared.message };
}

export async function processCashDrawerAction({
  action,
  amountInput,
  reason,
  drawer,
  context,
  syncMovement,
  enqueueMovement,
}: ProcessCashDrawerActionParams): Promise<ProcessCashDrawerActionResult> {
  if (action === "float") {
    const floatResult = recordCashDrawerAction({
      action,
      amountInput,
      reason,
      drawer,
    });
    return { ...floatResult, queued: false };
  }

  // Validate before syncing or mutating so a rejected movement never leaves an
  // unsynced entry in the local drawer.
  const { branchId, employeeId } = requireMovementContext(context);
  const prepared = prepareCashDrawerAction({ action, amountInput, reason });
  let queued = false;

  try {
    const sessionId = await context.resolveSessionId();
    const payload: CashMovementPayload = {
      branch_id: branchId,
      terminal_id: context.terminalId ?? null,
      session_id: sessionId,
      employee_id: employeeId,
      amount: prepared.amount,
      reason: prepared.reason || null,
    };
    await syncMovement(action, payload);
  } catch (error) {
    // A missing session is a permanent local condition, not an offline signal.
    if (
      error instanceof NoOpenRegisterSessionError ||
      !isRetryableTransportError(error)
    ) {
      throw error;
    }
    // Queue with a session_id of 0; the replay path re-resolves the open
    // session before posting.
    const payload: CashMovementPayload = {
      branch_id: branchId,
      terminal_id: context.terminalId ?? null,
      session_id: 0,
      employee_id: employeeId,
      amount: prepared.amount,
      reason: prepared.reason || null,
    };
    enqueueMovement({ action, payload });
    queued = true;
  }

  applyCashDrawerAction(action, prepared.amount, prepared.reason, drawer);
  return { amount: prepared.amount, message: prepared.message, queued };
}
