import React, { useMemo, useState } from 'react';
import { toast } from 'sonner';
import ScreenLayout from '../../layouts/ScreenLayout';
import { useCashDrawerStore } from '../../shared/stores/cashDrawerStore';
import { useCashMovementQueueStore } from '../../shared/stores/cashMovementQueueStore';
import { formatCurrency, formatDate } from '../../shared/utils/format';
import { processCashDrawerAction, type CashDrawerAction } from './cashDrawerWorkflow';
import { syncCashMovementWithApi } from './cashMovementQueueSync';
import { useTerminalContextStore } from '../../shared/stores/terminalContextStore';
import { useEmployeeSessionStore } from '../../shared/stores/employeeSessionStore';
import { resolveOpenSessionId } from '../close-session/closeSessionWorkflow';

const CashDropScreen: React.FC = () => {
  const movements = useCashDrawerStore((s) => s.movements);
  const openingFloat = useCashDrawerStore((s) => s.openingFloat);
  const openDrawer = useCashDrawerStore((s) => s.openDrawer);
  const cashDrop = useCashDrawerStore((s) => s.cashDrop);
  const paidOut = useCashDrawerStore((s) => s.paidOut);
  const paidIn = useCashDrawerStore((s) => s.paidIn);
  const expectedCashFn = useCashDrawerStore((s) => s.expectedCash);
  const expectedCash = useMemo(() => expectedCashFn(), [expectedCashFn, movements, openingFloat]);
  const enqueueMovement = useCashMovementQueueStore((s) => s.enqueue);
  const branchId = useTerminalContextStore((s) => s.locationId);
  const terminalId = useTerminalContextStore((s) => s.terminalId);
  const employee = useEmployeeSessionStore((s) => s.employee);
  const [action, setAction] = useState<CashDrawerAction>('drop');
  const [amount, setAmount] = useState('');
  const [reason, setReason] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const submit = async () => {
    if (submitting) return;
    setSubmitting(true);
    try {
      const result = await processCashDrawerAction({
        action,
        amountInput: amount,
        reason,
        drawer: { openDrawer, cashDrop, paidOut, paidIn },
        context: {
          branchId,
          terminalId,
          employeeId: Number(employee?.id) || null,
          resolveSessionId: resolveOpenSessionId,
        },
        syncMovement: syncCashMovementWithApi,
        enqueueMovement,
      });
      if (result.queued) {
        toast.warning(`${result.message}. Saved offline and will sync when connection returns.`);
      } else {
        toast.success(result.message);
      }
      setAmount('');
      setReason('');
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Unable to record cash movement.');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <ScreenLayout title="Cash Drop / Paid Out">
      <div className="p-5 max-w-3xl space-y-5">
        <div className="grid grid-cols-2 gap-3">
          <div className="p-4 rounded-xl bg-pos-card border border-pos-border">
            <div className="text-xs text-pos-muted uppercase">Opening float</div>
            <div className="text-2xl font-bold text-pos-text">{formatCurrency(openingFloat)}</div>
          </div>
          <div className="p-4 rounded-xl bg-pos-card border border-pos-border">
            <div className="text-xs text-pos-muted uppercase">Expected cash</div>
            <div className="text-2xl font-bold text-pos-text">{formatCurrency(expectedCash)}</div>
          </div>
        </div>
        <div className="rounded-xl bg-pos-card border border-pos-border p-4 space-y-3">
          <div className="grid grid-cols-4 gap-2">
            {(['float', 'drop', 'paid_out', 'paid_in'] as CashDrawerAction[]).map((a) => (
              <button
                key={a}
                onClick={() => setAction(a)}
                className={`h-11 rounded-lg text-xs font-semibold ${
                  action === a ? 'bg-pos-text text-pos-bg' : 'bg-white/8 text-pos-text'
                }`}
              >
                {a.replace('_', ' ')}
              </button>
            ))}
          </div>
          <input
            type="number"
            inputMode="decimal"
            placeholder="Amount"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            className="w-full h-12 px-3 rounded-lg bg-white/8 border border-pos-border text-pos-text outline-none"
          />
          {action !== 'float' && (
            <input
              placeholder="Reason"
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              className="w-full h-12 px-3 rounded-lg bg-white/8 border border-pos-border text-pos-text outline-none"
            />
          )}
          <button
            onClick={() => void submit()}
            disabled={submitting}
            className="w-full h-12 rounded-lg bg-pos-text text-pos-bg font-bold disabled:opacity-50"
          >
            {submitting ? 'Recording…' : 'Record'}
          </button>
        </div>
        <div>
          <h2 className="text-sm font-bold text-pos-text mb-2">Recent movements</h2>
          {movements.length === 0 ? (
            <p className="text-xs text-pos-muted">None yet</p>
          ) : (
            <ul className="space-y-1.5">
              {movements.slice(0, 20).map((m) => (
                <li
                  key={m.id}
                  className="flex items-center justify-between p-2 rounded-lg bg-white/5 border border-pos-border text-xs"
                >
                  <span className="uppercase text-pos-muted">{m.type.replace('_', ' ')}</span>
                  <span className="text-pos-text">{formatCurrency(m.amount)}</span>
                  <span className="text-pos-muted">{formatDate(m.at)}</span>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>
    </ScreenLayout>
  );
};

export default CashDropScreen;
