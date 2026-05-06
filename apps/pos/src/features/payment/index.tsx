import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { motion } from 'motion/react';
import { toast } from 'sonner';
import { unwrap } from '@voxire/api';
import POSLayout from '../../layouts/POSLayout';
import { useCartStore } from '../../shared/stores/cartStore';
import { useSettingsStore } from '../../shared/stores/settingsStore';
import { useEmployeeSessionStore } from '../../shared/stores/employeeSessionStore';
import { posService } from '../../shared/services/posService';
import { useOfflineQueueStore } from '../../shared/stores/offlineQueueStore';
import {
  buildCashTender,
  buildCheckoutOrder,
  processCheckout,
  type CheckoutPaymentMethod,
} from './checkoutWorkflow';
import { buildReceiptFromOrder } from './receiptArtifact';
import { useReceiptsStore } from '../../shared/stores/receiptsStore';
import { useCashDrawerStore } from '../../shared/stores/cashDrawerStore';
import { openConfiguredCashDrawer } from '../../shared/hardware/cashDrawerHardware';
import { resetAfterCheckout } from './checkoutState';

const fmt = (n: number): string =>
  new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(n);

const Payment: React.FC = () => {
  const navigate = useNavigate();
  const { items, total, clearCart, newOrder } = useCartStore();
  const taxRate = useSettingsStore((s) => s.taxRate);
  const printer = useSettingsStore((s) => s.printer);
  const employee = useEmployeeSessionStore((s) => s.employee);
  const enqueueOffline = useOfflineQueueStore((s) => s.enqueue);
  const addReceipt = useReceiptsStore((s) => s.add);
  const recordCashSale = useCashDrawerStore((s) => s.recordCashSale);
  const [method, setMethod] = useState<CheckoutPaymentMethod>('card');
  const [processing, setProcessing] = useState(false);
  const [cashAmount, setCashAmount] = useState('');

  const subtotal = total();
  const tax = subtotal * taxRate;
  const grandTotal = subtotal + tax;
  const cashTender = buildCashTender(cashAmount, grandTotal);

  const handleCharge = async (): Promise<void> => {
    setProcessing(true);
    try {
      const order = buildCheckoutOrder({
        id: `local-${Date.now()}`,
        items,
        subtotal,
        tax,
        total: grandTotal,
        employeeId: employee?.id,
        createdAt: new Date().toISOString(),
      });

      const result = await processCheckout({
        order,
        method,
        createOrder: async (payload) => unwrap(posService.createOrder(payload)),
        chargeOrder: async (orderId, payload) => {
          await unwrap(posService.chargeOrder(orderId, payload));
        },
        enqueueOffline,
        recordCashSale,
      });

      addReceipt(
        buildReceiptFromOrder({
          order,
          method,
          printedAt: new Date().toISOString(),
        })
      );
      resetAfterCheckout({ newOrder, clearCart });
      if (result.status === 'queued') {
        toast.warning('Cash order saved offline. It will sync when connection returns.');
      }
      if (method === 'cash') {
        const drawerResult = await openConfiguredCashDrawer({
          enabled: printer.cashDrawerKick,
          printerName: printer.name,
        });
        if (drawerResult.status === 'failed') {
          toast.warning(`Cash drawer did not open: ${drawerResult.error}`);
        }
      }
      navigate('/receipt');
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Payment failed. Please try again.');
    } finally {
      setProcessing(false);
    }
  };

  if (items.length === 0) {
    navigate('/');
    return null;
  }

  const methods: { id: CheckoutPaymentMethod; label: string }[] = [
    { id: 'card', label: '💳 Card' },
    { id: 'cash', label: '💵 Cash' },
    { id: 'qr', label: '📱 QR Pay' },
  ];

  return (
    <POSLayout>
      <div className="h-full flex items-center justify-center p-6">
        <div className="w-full max-w-md bg-pos-card border border-pos-border rounded-2xl p-6 space-y-5">
          <button
            className="text-sm text-pos-muted hover:text-pos-text transition-colors"
            onClick={() => navigate('/')}
          >
            ← Back to POS
          </button>
          <h1 className="text-xl font-bold text-pos-text">Payment</h1>

          {/* Summary */}
          <div className="space-y-2 text-sm">
            {items.map((item) => (
              <div key={item.id} className="flex justify-between text-pos-muted">
                <span>
                  {item.quantity}× {item.name}
                </span>
                <span>{fmt(item.price * item.quantity)}</span>
              </div>
            ))}
            <div className="h-px bg-pos-border my-2" />
            <div className="flex justify-between text-pos-muted">
              <span>Subtotal</span>
              <span>{fmt(subtotal)}</span>
            </div>
            <div className="flex justify-between text-pos-muted">
              <span>Tax ({(taxRate * 100).toFixed(0)}%)</span>
              <span>{fmt(tax)}</span>
            </div>
            <div className="flex justify-between font-bold text-pos-text text-base pt-1">
              <span>Total</span>
              <span>{fmt(grandTotal)}</span>
            </div>
          </div>

          {/* Method selector */}
          <div className="flex gap-2">
            {methods.map((m) => (
              <button
                key={m.id}
                className={`flex-1 py-2.5 rounded-xl text-sm font-medium border transition-colors ${
                  method === m.id
                    ? 'bg-pos-text text-pos-bg border-transparent'
                    : 'border-pos-border text-pos-muted hover:border-white/20'
                }`}
                onClick={() => setMethod(m.id)}
              >
                {m.label}
              </button>
            ))}
          </div>

          {method === 'card' && (
            <p className="text-sm text-pos-muted text-center">
              Present card to terminal or tap to pay.
            </p>
          )}

          {method === 'cash' && (
            <div className="space-y-2">
              <label className="text-xs font-semibold uppercase tracking-wide text-pos-muted">
                Amount received
              </label>
              <input
                className="w-full bg-pos-input border border-pos-border rounded-xl px-4 py-3 text-pos-text text-sm outline-none focus:border-white/20 transition-colors"
                type="number"
                step="0.01"
                min={grandTotal}
                value={cashAmount}
                onChange={(e) => setCashAmount(e.target.value)}
                placeholder={fmt(grandTotal)}
              />
              {cashTender.change !== null && cashTender.change >= 0 && (
                <motion.div
                  className="text-center text-sm text-success font-medium"
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                >
                  Change: {fmt(cashTender.change)}
                </motion.div>
              )}
            </div>
          )}

          {method === 'qr' && (
            <div className="flex justify-center">
              <div className="w-40 h-40 rounded-2xl bg-white/8 border border-pos-border flex flex-col items-center justify-center gap-1">
                <span className="text-3xl">⬜</span>
                <span className="text-xs text-pos-muted">Scan to pay</span>
              </div>
            </div>
          )}

          <button
            className="w-full py-3.5 rounded-xl bg-pos-text text-pos-bg font-bold text-sm hover:opacity-90 disabled:opacity-40 transition-opacity"
            onClick={handleCharge}
            disabled={
              processing ||
              (method === 'cash' && !cashTender.canCharge)
            }
          >
            {processing ? (
              <span className="inline-flex items-center gap-2">
                <span className="w-4 h-4 border-2 border-pos-bg border-t-transparent rounded-full animate-spin-slow" />
                Processing…
              </span>
            ) : (
              `Confirm Payment · ${fmt(grandTotal)}`
            )}
          </button>
        </div>
      </div>
    </POSLayout>
  );
};

export default Payment;
