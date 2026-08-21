import { createOneKitClient } from '@voxire/api';
import { OrderDiscountServiceClient } from '@voxire/voxie/pos/orders/discounts/v1/client';
import React, { useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { toast } from 'sonner';
import ScreenLayout from '../../layouts/ScreenLayout';
import ManagerRequired from '../../shared/components/ManagerRequired';
import { useCapabilities } from '../../shared/hooks/useCapabilities';
import { formatCurrency } from '../../shared/utils/format';
import { buildDiscountRequest, type DiscountMode } from './discountWorkflow';

const orderDiscountClient = createOneKitClient(OrderDiscountServiceClient, 'pos');

const DiscountScreen: React.FC = () => {
  const { orderId } = useParams<{ orderId: string }>();
  const navigate = useNavigate();
  const { canApplyDiscount, maxDiscountPct } = useCapabilities();
  const [mode, setMode] = useState<DiscountMode>('percent');
  const [value, setValue] = useState('');
  const [reason, setReason] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const submit = async () => {
    if (submitting) return;
    setSubmitting(true);
    try {
      const request = buildDiscountRequest({ orderId, mode, valueInput: value, reason });
      // Enforce the operator's discount ceiling for percent discounts.
      if (
        mode === 'percent' &&
        maxDiscountPct > 0 &&
        (request.payload.discountValue ?? 0) > maxDiscountPct
      ) {
        toast.error(`Your discount limit is ${maxDiscountPct}%`);
        return;
      }
      await orderDiscountClient.discountOrder(request.payload);
      toast.success('Discount applied');
      navigate(-1);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Could not apply discount');
    } finally {
      setSubmitting(false);
    }
  };

  const discountPreview = Number(value);
  const canSubmit = Number.isFinite(discountPreview) && discountPreview > 0;

  return (
    <ScreenLayout title="Apply Discount" subtitle={`Order #${orderId ?? 'current'}`}>
      <div className="p-5 max-w-md space-y-5">
        {!canApplyDiscount ? (
          <ManagerRequired action="apply discounts" />
        ) : (
          <>
            <div className="grid grid-cols-2 gap-2">
              <button
                onClick={() => setMode('percent')}
                className={`h-14 rounded-lg font-semibold ${
                  mode === 'percent' ? 'bg-pos-text text-pos-bg' : 'bg-white/8 text-pos-text'
                }`}
              >
                Percent
              </button>
              <button
                onClick={() => setMode('flat')}
                className={`h-14 rounded-lg font-semibold ${
                  mode === 'flat' ? 'bg-pos-text text-pos-bg' : 'bg-white/8 text-pos-text'
                }`}
              >
                Flat amount
              </button>
            </div>
            <div>
              <label htmlFor="discount-value" className="text-xs text-pos-muted">
                {mode === 'percent' ? 'Discount %' : 'Amount'}
              </label>
              <input
                id="discount-value"
                type="number"
                inputMode="decimal"
                value={value}
                onChange={(e) => setValue(e.target.value)}
                placeholder={mode === 'percent' ? '10' : '5.00'}
                className="w-full h-12 mt-1 px-4 rounded-lg bg-white/8 border border-pos-border text-pos-text outline-none focus:border-white/25"
              />
            </div>
            <div>
              <label htmlFor="discount-reason" className="text-xs text-pos-muted">
                Reason
              </label>
              <input
                id="discount-reason"
                type="text"
                value={reason}
                onChange={(e) => setReason(e.target.value)}
                placeholder="e.g. manager approval, comped item"
                maxLength={200}
                className="w-full h-12 mt-1 px-4 rounded-lg bg-white/8 border border-pos-border text-pos-text outline-none focus:border-white/25"
              />
            </div>
            <button
              onClick={() => void submit()}
              disabled={!canSubmit || submitting || (mode === 'percent' && discountPreview > 100)}
              className="w-full h-14 rounded-xl bg-pos-text text-pos-bg font-bold disabled:opacity-50"
            >
              {submitting ? 'Applying…' : 'Apply'}{' '}
              {mode === 'percent'
                ? `${Number.isFinite(discountPreview) ? discountPreview : 0}%`
                : formatCurrency(Number.isFinite(discountPreview) ? discountPreview : 0)}
            </button>
          </>
        )}
      </div>
    </ScreenLayout>
  );
};

export default DiscountScreen;
