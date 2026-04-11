import React, { useState } from 'react';
import { toast } from 'sonner';
import ScreenLayout from '../../layouts/ScreenLayout';
import { useCashDrawerStore } from '../../shared/stores/cashDrawerStore';
import { formatCurrency, formatDate } from '../../shared/utils/format';

type Action = 'drop' | 'paid_out' | 'paid_in' | 'float';

const CashDropScreen: React.FC = () => {
  const {
    movements,
    openingFloat,
    openDrawer,
    cashDrop,
    paidOut,
    paidIn,
    expectedCash,
  } = useCashDrawerStore();
  const [action, setAction] = useState<Action>('drop');
  const [amount, setAmount] = useState('');
  const [reason, setReason] = useState('');

  const submit = () => {
    const value = Number(amount || 0);
    if (value <= 0) {
      toast.error('Enter an amount');
      return;
    }
    if (action === 'float') {
      openDrawer(value);
      toast.success(`Opening float set to ${formatCurrency(value)}`);
    } else if (action === 'drop') {
      cashDrop(value, reason);
      toast.success(`Cash drop recorded`);
    } else if (action === 'paid_out') {
      paidOut(value, reason);
      toast.success(`Paid out`);
    } else if (action === 'paid_in') {
      paidIn(value, reason);
      toast.success(`Paid in`);
    }
    setAmount('');
    setReason('');
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
            <div className="text-2xl font-bold text-pos-text">{formatCurrency(expectedCash())}</div>
          </div>
        </div>
        <div className="rounded-xl bg-pos-card border border-pos-border p-4 space-y-3">
          <div className="grid grid-cols-4 gap-2">
            {(['float', 'drop', 'paid_out', 'paid_in'] as Action[]).map((a) => (
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
            onClick={submit}
            className="w-full h-12 rounded-lg bg-pos-text text-pos-bg font-bold"
          >
            Record
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
