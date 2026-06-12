import { unwrap } from '@voxire/api';
import React, { useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { toast } from 'sonner';
import ScreenLayout from '../../layouts/ScreenLayout';
import ConfirmModal from '../../shared/components/ConfirmModal';
import { posService } from '../../shared/services/api/posService';
import { useCashDrawerStore } from '../../shared/stores/cashDrawerStore';
import { useReceiptsStore } from '../../shared/stores/receiptsStore';
import { formatCurrency } from '../../shared/utils/format';
import { processRefund } from './refundWorkflow';

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

  const submit = async () => {
    setBusy(true);
    try {
      await processRefund({
        orderId,
        fullRefund,
        amountInput: amount,
        reason,
        receipt,
        refundOrder: async (id, payload) => {
          await unwrap(posService.refundOrder(id, payload));
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
