import { unwrap } from '@voxire/api';
import React, { useEffect, useMemo, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { toast } from 'sonner';
import { Loader2 } from 'lucide-react';
import ScreenLayout from '../../layouts/ScreenLayout';
import PaymentMethodTile from '../../shared/components/PaymentMethodTile';
import TipSelector from '../../shared/components/TipSelector';
import { formatCurrency } from '../../shared/utils/format';
import type { PaymentEntry, PaymentMethod } from '../../shared/types/pos';
import { posService } from '../../shared/services/api/posService';
import { finalizeMultiPayment } from './multiPaymentWorkflow';
import { useSettingsStore } from '../../shared/stores/settingsStore';

interface OrderTotals {
  subtotal: number;
  tax: number;
  total: number;
}

const MultiPaymentScreen: React.FC = () => {
  const { orderId } = useParams<{ orderId: string }>();
  const navigate = useNavigate();
  // MultiPayment pays a PERSISTED order (route /payment/:orderId), so totals
  // come from that order — never the live cart, which is unrelated.
  const [order, setOrder] = useState<OrderTotals | null>(null);
  const [loadingOrder, setLoadingOrder] = useState(true);
  const [tip, setTip] = useState(0);
  const [payments, setPayments] = useState<PaymentEntry[]>([]);
  const [method, setMethod] = useState<PaymentMethod>('cash');
  const [amount, setAmount] = useState('');
  const [busy, setBusy] = useState(false);
  const tipsEnabled = useSettingsStore((s) => s.features.tips);

  useEffect(() => {
    let active = true;
    void (async () => {
      const id = Number(orderId);
      if (!Number.isInteger(id) || id <= 0) {
        if (active) setLoadingOrder(false);
        return;
      }
      try {
        const res = await unwrap(posService.getOrder(id));
        const o = ((res as { order?: Record<string, unknown> })?.order ??
          res) as Record<string, unknown>;
        if (active) {
          setOrder({
            subtotal: Number(o.subtotal ?? 0),
            tax: Number(o.taxAmount ?? o.tax ?? 0),
            total: Number(o.total ?? 0),
          });
        }
      } catch {
        if (active) toast.error('Could not load the order to charge');
      } finally {
        if (active) setLoadingOrder(false);
      }
    })();
    return () => {
      active = false;
    };
  }, [orderId]);

  const subtotal = order?.subtotal ?? 0;
  const tax = order?.tax ?? 0;
  const appliedTip = tipsEnabled ? tip : 0;
  const total = (order?.total ?? 0) + appliedTip;
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
        tip: appliedTip,
        total,
        chargeOrder: async (id, payload) => {
          await unwrap(posService.chargeOrder(id, payload));
        },
      });
      toast.success('Payment complete');
      navigate('/receipts');
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Payment failed. Please try again.');
    } finally {
      setBusy(false);
    }
  };

  if (loadingOrder) {
    return (
      <ScreenLayout title="Payment" subtitle={`Order #${orderId ?? ''}`}>
        <div className="flex justify-center py-16">
          <Loader2 className="w-7 h-7 animate-spin text-pos-muted" />
        </div>
      </ScreenLayout>
    );
  }

  if (!order) {
    return (
      <ScreenLayout title="Payment" subtitle={`Order #${orderId ?? ''}`}>
        <p className="p-5 text-sm text-pos-muted">
          Couldn't load this order. Go back and reopen it from the order list.
        </p>
      </ScreenLayout>
    );
  }

  return (
    <ScreenLayout title="Payment" subtitle={`Order #${orderId ?? ''}`}>
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
            {tipsEnabled && (
              <div className="flex justify-between text-sm text-pos-muted">
                <span>Tip</span>
                <span>{formatCurrency(appliedTip)}</span>
              </div>
            )}
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
          {tipsEnabled && <TipSelector subtotal={subtotal + tax} onTipChange={setTip} />}
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
