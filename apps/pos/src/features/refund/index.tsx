import { createOneKitClient } from '@voxire/api';
import { OrderRefundServiceClient } from '@voxire/voxie/pos/orders/refunds/v1/client';
import { OrderServiceClient } from '@voxire/voxie/pos/orders/v1/client';
import type { Order } from '@voxire/voxie/pos/orders/v1/types';
import React, { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { toast } from 'sonner';
import ScreenLayout from '../../layouts/ScreenLayout';
import ConfirmModal from '../../shared/components/ConfirmModal';
import ManagerRequired from '../../shared/components/ManagerRequired';
import { useCapabilities } from '../../shared/hooks/useCapabilities';
import { useCashDrawerStore } from '../../shared/stores/cashDrawerStore';
import { useReceiptsStore } from '../../shared/stores/receiptsStore';
import { formatCurrency } from '../../shared/utils/format';
import { processRefund } from './refundWorkflow';

const orderRefundClient = createOneKitClient(OrderRefundServiceClient, 'pos');
const orderClient = createOneKitClient(OrderServiceClient, 'pos');

const REASONS = [
  'Customer complaint',
  'Wrong item',
  'Quality issue',
  'Service delay',
  'Manager decision',
  'Other',
];

const RefundScreen: React.FC = () => {
  const { orderId } = useParams<{ orderId: string }>();
  const navigate = useNavigate();
  const [amount, setAmount] = useState('');
  const [reason, setReason] = useState(REASONS[0]);
  const [fullRefund, setFullRefund] = useState(true);
  const [confirm, setConfirm] = useState(false);
  const [busy, setBusy] = useState(false);
  const receipt = useReceiptsStore((s) => (orderId ? s.byOrder(orderId) : undefined));
  const recordCashRefund = useCashDrawerStore((s) => s.recordCashRefund);
  const { canRefund } = useCapabilities();
  // Server order data backs the full-refund total and the default refund
  // tender; local receipts rarely exist for server order ids.
  const [order, setOrder] = useState<Order | null>(null);
  const [refundMethod, setRefundMethod] = useState<'cash' | 'other'>('cash');

  useEffect(() => {
    const id = Number(orderId);
    if (!orderId || !Number.isInteger(id) || id <= 0) return;
    let active = true;
    orderClient
      .getOrderDetail({ orderId: String(id) })
      .then((res) => {
        if (!active) return;
        setOrder(res.order ?? null);
        if (res.order?.paymentMethod && res.order.paymentMethod !== 'cash') {
          setRefundMethod('other');
        }
      })
      .catch(() => {
        // Total falls back to the local receipt; the workflow throws an
        // honest error when neither source has a total.
      });
    return () => {
      active = false;
    };
  }, [orderId]);

  const submit = async () => {
    setBusy(true);
    try {
      await processRefund({
        orderId,
        fullRefund,
        amountInput: amount,
        reason,
        orderTotal: order?.total,
        receipt,
        refundMethod,
        refundOrder: async (payload) => {
          await orderRefundClient.processRefund(payload);
        },
        recordCashRefund,
      });
      toast.success('Refund processed');
      navigate('/receipts');
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Refund failed');
    } finally {
      setBusy(false);
      setConfirm(false);
    }
  };

  if (!canRefund) {
    return (
      <ScreenLayout title="Refund" subtitle={`Order #${orderId ?? 'current'}`}>
        <ManagerRequired action="issue refunds" />
      </ScreenLayout>
    );
  }

  return (
    <ScreenLayout title="Refund" subtitle={`Order #${orderId ?? 'current'}`}>
      <div className="p-5 max-w-xl space-y-5">
        <div className="grid grid-cols-2 gap-2">
          <button
            onClick={() => setFullRefund(true)}
            className={`h-14 rounded-lg font-semibold ${
              fullRefund ? 'bg-pos-text text-pos-bg' : 'bg-white/8 text-pos-text'
            }`}
          >
            Full Refund
          </button>
          <button
            onClick={() => setFullRefund(false)}
            className={`h-14 rounded-lg font-semibold ${
              !fullRefund ? 'bg-pos-text text-pos-bg' : 'bg-white/8 text-pos-text'
            }`}
          >
            Partial Refund
          </button>
        </div>
        {!fullRefund && (
          <div className="space-y-1">
            <label className="text-xs text-pos-muted">Amount</label>
            <input
              type="number"
              inputMode="decimal"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              placeholder="0.00"
              className="w-full h-12 px-4 rounded-lg bg-white/8 border border-pos-border text-pos-text outline-none focus:border-white/25"
            />
          </div>
        )}
        <div className="space-y-1">
          <label className="text-xs text-pos-muted">Reason</label>
          <select
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            className="w-full h-12 px-4 rounded-lg bg-white/8 border border-pos-border text-pos-text outline-none"
          >
            {REASONS.map((r) => (
              <option key={r}>{r}</option>
            ))}
          </select>
        </div>
        <div className="space-y-1">
          <label className="text-xs text-pos-muted">Refund method</label>
          <div className="grid grid-cols-2 gap-2">
            <button
              type="button"
              onClick={() => setRefundMethod('cash')}
              className={`h-12 rounded-lg font-semibold ${
                refundMethod === 'cash' ? 'bg-pos-text text-pos-bg' : 'bg-white/8 text-pos-text'
              }`}
            >
              Cash
            </button>
            <button
              type="button"
              onClick={() => setRefundMethod('other')}
              className={`h-12 rounded-lg font-semibold ${
                refundMethod === 'other' ? 'bg-pos-text text-pos-bg' : 'bg-white/8 text-pos-text'
              }`}
            >
              Card / Other
            </button>
          </div>
          {refundMethod !== 'cash' && (
            <p className="text-[11px] text-pos-muted">
              Non-cash refunds do not adjust the cash drawer.
            </p>
          )}
        </div>
        <button
          onClick={() => setConfirm(true)}
          disabled={!fullRefund && !Number(amount)}
          className="w-full h-14 rounded-xl bg-destructive text-white font-bold text-base disabled:opacity-50"
        >
          {fullRefund ? 'Issue Full Refund' : `Issue ${formatCurrency(Number(amount) || 0)}`}
        </button>
      </div>
      <ConfirmModal
        open={confirm}
        title="Confirm refund"
        message="This action cannot be undone."
        confirmText={busy ? 'Refunding…' : 'Refund'}
        danger
        onCancel={() => setConfirm(false)}
        onConfirm={submit}
      />
    </ScreenLayout>
  );
};

export default RefundScreen;
