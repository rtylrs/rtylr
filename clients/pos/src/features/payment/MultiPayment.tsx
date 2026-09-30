import { createOneKitClient } from '@voxire/api';
import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { toast } from 'sonner';
import { Loader2 } from 'lucide-react';
import ScreenLayout from '../../layouts/ScreenLayout';
import PaymentMethodTile from '../../shared/components/PaymentMethodTile';
import TipSelector from '../../shared/components/TipSelector';
import { formatCurrency } from '../../shared/utils/format';
import type { PaymentEntry, PaymentMethod } from '../../shared/types/pos';
import type { Order, OrderItem } from '@voxire/voxie/pos/orders/v1/types';
import { OrderServiceClient } from '@voxire/voxie/pos/orders/v1/client';
import { OrderPaymentServiceClient } from '@voxire/voxie/pos/orders/payments/v1/client';
import { finalizeMultiPayment } from './multiPaymentWorkflow';
import { useSettingsStore } from '../../shared/stores/settingsStore';
import { isOrderFinalized } from '../../shared/utils/orderStatus';
import type { OrderStatus } from '../../shared/types/pos';

const orderClient = createOneKitClient(OrderServiceClient, 'pos');
const orderPaymentClient = createOneKitClient(OrderPaymentServiceClient, 'pos');

function lineTotal(item: OrderItem): number {
  return item.total || (item.unitPrice ?? 0) * (item.quantity ?? 0);
}

/** Settled orders ('completed' is a legacy server value) must never be re-charged. */
function isSettled(status: string | undefined): boolean {
  return isOrderFinalized((status ?? '') as OrderStatus) || status === 'completed';
}

const MultiPaymentScreen: React.FC = () => {
  const { orderId } = useParams<{ orderId: string }>();
  const navigate = useNavigate();
  // MultiPayment pays a PERSISTED order (route /payment/:orderId), so totals
  // come from that order — never the live cart, which is unrelated.
  const [order, setOrder] = useState<Order | null>(null);
  const [loadingOrder, setLoadingOrder] = useState(true);
  const [orderError, setOrderError] = useState<string | null>(null);
  const [tip, setTip] = useState(0);
  const [payments, setPayments] = useState<PaymentEntry[]>([]);
  const [method, setMethod] = useState<PaymentMethod>('cash');
  const [amount, setAmount] = useState('');
  const [busy, setBusy] = useState(false);
  // Same reason as the checkout screen: `busy` disables the button on the next
  // render only, so a second tap in the same tick would charge twice.
  const finalizing = useRef(false);
  const tipsEnabled = useSettingsStore((s) => s.features.tips);

  const loadOrder = useCallback(async (canApply: () => boolean = () => true): Promise<void> => {
    const id = Number(orderId);
    setLoadingOrder(true);
    setOrderError(null);

    if (!Number.isInteger(id) || id <= 0) {
      if (canApply()) {
        setOrder(null);
        setOrderError('This order link is invalid.');
        setLoadingOrder(false);
      }
      return;
    }

    try {
      const res = await orderClient.getOrderDetail({ orderId: String(id) });
      const order = res.order ?? null;
      if (canApply()) {
        setOrder(order);
        setOrderError(order ? null : 'The order response was empty.');
      }
    } catch (error) {
      const message =
        error instanceof Error ? error.message : 'Could not load the order to charge';
      if (canApply()) {
        setOrder(null);
        setOrderError(message);
        toast.error(message);
      }
    } finally {
      if (canApply()) setLoadingOrder(false);
    }
  }, [orderId]);

  useEffect(() => {
    let active = true;
    void loadOrder(() => active);
    return () => {
      active = false;
    };
  }, [loadOrder]);

  const subtotal = order?.subtotal ?? 0;
  const tax = order?.taxAmount ?? 0;
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
    if (order && isSettled(order.status)) {
      toast.error(
        `Order #${order.orderNumber || order.id} is already ${order.status} — no charge was made`,
      );
      return;
    }
    if (paid + 0.001 < total) {
      toast.error('Remaining balance not covered');
      return;
    }
    if (paid > total + 0.001) {
      toast.error('Collected payments exceed the amount due');
      return;
    }
    if (finalizing.current) return;
    finalizing.current = true;
    setBusy(true);
    try {
      await finalizeMultiPayment({
        orderId,
        payments,
        chargeOrder: async ({ idempotencyKey, ...payload }) => {
          // Required on this route: the server answers 428 without it.
          await orderPaymentClient.processPayment(payload, {
            headers: { 'Idempotency-Key': idempotencyKey },
          });
        },
      });
      toast.success('Payment complete');
      navigate('/receipts');
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Payment failed. Please try again.');
    } finally {
      setBusy(false);
      finalizing.current = false;
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
        <div className="p-5 space-y-4">
          <div className="rounded-xl border border-pos-border bg-pos-card p-4">
            <p className="text-sm font-semibold text-pos-text">Couldn't load this order.</p>
            <p className="mt-1 text-sm text-pos-muted">
              {orderError ?? 'Go back and reopen it from the order list.'}
            </p>
            <div className="mt-4 flex flex-wrap gap-2">
              <button
                type="button"
                onClick={() => void loadOrder()}
                className="h-10 px-4 rounded-lg bg-pos-text text-pos-bg text-sm font-bold"
              >
                Try again
              </button>
              <button
                type="button"
                onClick={() => navigate('/orders')}
                className="h-10 px-4 rounded-lg bg-white/8 text-pos-text text-sm font-bold hover:bg-white/12"
              >
                Back to orders
              </button>
            </div>
          </div>
        </div>
      </ScreenLayout>
    );
  }

  return (
    <ScreenLayout title="Payment" subtitle={`Order #${orderId ?? ''}`}>
      <div className="w-full p-5 grid grid-cols-1 lg:grid-cols-2 gap-5">
        <div className="space-y-5">
          <div className="rounded-xl bg-pos-card border border-pos-border p-4">
            <div className="flex items-center justify-between gap-3 mb-3">
              <div>
                <p className="text-sm font-bold text-pos-text">Items purchased</p>
                <p className="text-[11px] text-pos-muted">
                  {(order.items ?? []).length} line{(order.items ?? []).length === 1 ? '' : 's'}
                </p>
              </div>
              <span className="text-xs text-pos-muted">#{order.orderNumber || order.id}</span>
            </div>
            {(order.items ?? []).length > 0 ? (
              <ul className="space-y-2">
                {(order.items ?? []).map((item) => (
                  <li
                    key={item.id || `${item.name}-${item.quantity}-${item.unitPrice}`}
                    className="rounded-lg bg-white/6 border border-pos-border p-3"
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div className="min-w-0">
                        <p className="text-sm font-semibold text-pos-text truncate">
                          {item.quantity}x {item.name}
                        </p>
                        <p className="text-[11px] text-pos-muted">
                          {formatCurrency(item.unitPrice ?? 0)} each
                          {item.sku ? ` · ${item.sku}` : ''}
                        </p>
                        {item.notes && (
                          <p className="text-[11px] italic text-pos-muted mt-1">
                            Note: {item.notes}
                          </p>
                        )}
                      </div>
                      <span className="text-sm font-bold text-pos-text shrink-0">
                        {formatCurrency(lineTotal(item))}
                      </span>
                    </div>
                  </li>
                ))}
              </ul>
            ) : (
              <p className="rounded-lg bg-white/6 border border-pos-border p-3 text-sm text-pos-muted">
                No items were returned for this order.
              </p>
            )}
          </div>

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
            disabled={busy || remaining > 0.001 || isSettled(order.status)}
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
