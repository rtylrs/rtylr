import { unwrap } from '@voxire/api';
import React, { useMemo, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { toast } from 'sonner';
import ScreenLayout from '../../layouts/ScreenLayout';
import PaymentMethodTile from '../../shared/components/PaymentMethodTile';
import TipSelector from '../../shared/components/TipSelector';
import { formatCurrency } from '../../shared/utils/format';
import { useCartStore } from '../../shared/stores/cartStore';
import { useSettingsStore } from '../../shared/stores/settingsStore';
import type { PaymentEntry, PaymentMethod } from '../../shared/types/pos';
import { posService } from '../../shared/services/api/posService';
import { finalizeMultiPayment } from './multiPaymentWorkflow';

const MultiPaymentScreen: React.FC = () => {
  const { orderId } = useParams<{ orderId: string }>();
  const navigate = useNavigate();
  const subtotal = useCartStore((s) => s.total());
  const clear = useCartStore((s) => s.clearCart);
  const taxRate = useSettingsStore((s) => s.taxRate);
  const tax = subtotal * taxRate;
  const [tip, setTip] = useState(0);
  const [payments, setPayments] = useState<PaymentEntry[]>([]);
  const [method, setMethod] = useState<PaymentMethod>('cash');
  const [amount, setAmount] = useState('');
  const [busy, setBusy] = useState(false);

  const total = subtotal + tax + tip;
  const paid = payments.reduce((sum, p) => sum + p.amount, 0);
  const remaining = useMemo(() => Math.max(0, total - paid), [total, paid]);

  const addPayment = () => {
    const value = Number(amount || remaining);
    if (value <= 0) return;
    setPayments((prev) => [
      ...prev,
      {
        id: 'p_' + Math.random().toString(36).slice(2, 8),
        method,
        amount: Math.min(value, remaining),
        createdAt: new Date().toISOString(),
      },
    ]);
    setAmount('');
  };

  const removePayment = (id: string) => setPayments((prev) => prev.filter((p) => p.id !== id));

  const finalize = async () => {
    if (remaining > 0.001) {
      toast.error('Remaining balance not covered');
      return;
    }
    setBusy(true);
    try {
      await finalizeMultiPayment({
        orderId,
        payments,
        tip,
        total,
        chargeOrder: async (id, payload) => {
          await unwrap(posService.chargeOrder(id, payload));
        },
      });
      toast.success('Payment complete');
      clear();
      navigate('/receipts');
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Payment failed. Please try again.');
    } finally {
      setBusy(false);
    }
  };

  return (
    <ScreenLayout title="Payment" subtitle={`Order #${orderId ?? 'current'}`}>
      <div className="p-5 grid grid-cols-1 lg:grid-cols-2 gap-5 max-w-6xl">
        <div className="space-y-5">
          <div className="rounded-xl bg-pos-card border border-pos-border p-4">
            <div className="flex justify-between text-sm text-pos-muted">
              <span>Subtotal</span>
              <span>{formatCurrency(subtotal)}</span>
            </div>
            <div className="flex justify-between text-sm text-pos-muted">
              <span>Tax</span>
              <span>{formatCurrency(tax)}</span>
            </div>
            <div className="flex justify-between text-sm text-pos-muted">
              <span>Tip</span>
              <span>{formatCurrency(tip)}</span>
            </div>
            <div className="flex justify-between pt-2 mt-2 border-t border-pos-border">
              <span className="text-base font-bold text-pos-text">Total</span>
              <span className="text-base font-bold text-pos-text">{formatCurrency(total)}</span>
            </div>
            <div className="flex justify-between text-sm text-emerald-300">
              <span>Paid</span>
              <span>{formatCurrency(paid)}</span>
            </div>
            <div className="flex justify-between text-base font-bold text-amber-300 pt-1">
              <span>Remaining</span>
              <span>{formatCurrency(remaining)}</span>
            </div>
          </div>
          <TipSelector subtotal={subtotal + tax} onTipChange={setTip} />
        </div>

        <div className="space-y-4">
          <div className="grid grid-cols-2 gap-3">
            {(['cash', 'card', 'gift_card', 'other'] as PaymentMethod[]).map((m) => (
              <PaymentMethodTile
                key={m}
                method={m}
                active={method === m}
                onClick={() => setMethod(m)}
              />
            ))}
          </div>
          <div className="flex gap-2">
            <input
              type="number"
              inputMode="decimal"
              placeholder={remaining.toFixed(2)}
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              className="flex-1 h-12 px-4 rounded-lg bg-white/8 border border-pos-border text-pos-text outline-none focus:border-white/25"
            />
            <button
              onClick={addPayment}
              className="h-12 px-5 rounded-lg bg-white/10 text-pos-text font-semibold hover:bg-white/20"
            >
              Add Payment
            </button>
          </div>
          <ul className="space-y-2">
            {payments.map((p) => (
              <li
                key={p.id}
                className="flex items-center justify-between p-3 rounded-lg bg-white/6 border border-pos-border"
              >
                <span className="text-sm text-pos-text capitalize">{p.method}</span>
                <span className="text-sm font-semibold text-pos-text">
                  {formatCurrency(p.amount)}
                </span>
                <button onClick={() => removePayment(p.id)} className="text-xs text-destructive">
                  Remove
                </button>
              </li>
            ))}
          </ul>
          <button
            onClick={finalize}
            disabled={busy || remaining > 0.001}
            className="w-full h-14 rounded-xl bg-pos-text text-pos-bg font-bold disabled:opacity-50"
          >
            {busy ? 'Processing…' : `Finalize ${formatCurrency(total)}`}
          </button>
        </div>
      </div>
    </ScreenLayout>
  );
};

export default MultiPaymentScreen;
