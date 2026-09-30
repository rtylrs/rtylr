import { createOneKitClient } from '@voxire/api';
import { OrderSplitServiceClient } from '@voxire/voxie/pos/orders/splits/v1/client';
import React, { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate, useParams } from 'react-router-dom';
import { toast } from 'sonner';
import ScreenLayout from '../../layouts/ScreenLayout';
import { formatCurrency } from '../../shared/utils/format';
import { parseLocalizedAmount, sanitiseAmountInput } from '../../shared/utils/amount';
import { useCartStore } from '../../shared/stores/cartStore';
import { buildSplitBillRequest } from './splitBillWorkflow';

const orderSplitClient = createOneKitClient(OrderSplitServiceClient, 'pos');

type SplitMode = 'seat' | 'amount' | 'items';

const SplitBillScreen: React.FC = () => {
  const { orderId } = useParams<{ orderId: string }>();
  const navigate = useNavigate();
  const { i18n } = useTranslation();
  const locale = i18n?.resolvedLanguage || i18n?.language || 'en';
  const items = useCartStore((s) => s.items);
  const total = useCartStore((s) => s.total());
  const serverOrderId = useCartStore((s) => s.serverOrderId);
  const [mode, setMode] = useState<SplitMode>('seat');
  const [seats, setSeats] = useState(2);
  const [assignments, setAssignments] = useState<Record<string, number>>({});
  // Raw per-split inputs; empty entries fall back to the even split of the
  // CURRENT total so the fields never go stale after cart changes.
  const [amountInputs, setAmountInputs] = useState<string[]>([]);
  const [confirming, setConfirming] = useState(false);

  const evenSplit = total / seats;

  const amountSplits = useMemo(
    () =>
      amountInputs.length === seats
        ? amountInputs
        : Array.from({ length: seats }, () => evenSplit.toFixed(2)),
    [amountInputs, seats, evenSplit]
  );

  const amountSum = useMemo(
    () => amountSplits.reduce((sum, raw) => sum + parseLocalizedAmount(raw, locale), 0),
    [amountSplits, locale]
  );
  const amountMatchesTotal = total > 0 && Math.abs(amountSum - total) < 0.005;

  const seatTotals = useMemo(() => {
    const result: number[] = Array.from({ length: seats }, () => 0);
    items.forEach((item) => {
      const seat = assignments[item.id] ?? 0;
      result[seat] = (result[seat] ?? 0) + item.price * item.quantity;
    });
    return result;
  }, [items, assignments, seats]);

  const setSplitCount = (n: number) => {
    setSeats(n);
    setAmountInputs([]);
  };

  const updateAmount = (idx: number, raw: string) => {
    const text = sanitiseAmountInput(raw, locale);
    setAmountInputs(amountSplits.map((v, i) => (i === idx ? text : v)));
  };

  const confirm = async (): Promise<void> => {
    if (mode !== 'seat') {
      toast.error('Only seat-based splits can be confirmed right now.');
      return;
    }
    let request: ReturnType<typeof buildSplitBillRequest>;
    try {
      request = buildSplitBillRequest({
        routeOrderId: orderId,
        serverOrderId,
        seats,
        items,
        assignments,
      });
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to split order');
      return;
    }

    if (request.localOnly) {
      // Nothing is persisted for a purely local cart - do not pretend success.
      toast.warning('Sync this order with the server before splitting it.');
      return;
    }
    setConfirming(true);
    try {
      await orderSplitClient.splitOrder(request.payload);
      toast.success(`Split into ${seats} ticket${seats === 1 ? '' : 's'}`);
      navigate(-1);
    } catch {
      toast.error('Failed to split order');
    } finally {
      setConfirming(false);
    }
  };

  return (
    <ScreenLayout title="Split Payment" subtitle={`Order #${orderId ?? 'current'}`}>
      <div className="p-5 space-y-5 max-w-5xl">
        <div className="flex flex-wrap gap-2">
          {(['seat', 'amount', 'items'] as SplitMode[]).map((m) => (
            <button
              key={m}
              onClick={() => setMode(m)}
              className={`h-11 px-5 rounded-lg text-sm font-semibold ${
                mode === m
                  ? 'bg-pos-text text-pos-bg'
                  : 'bg-white/8 text-pos-text hover:bg-white/14'
              }`}
            >
              By {m}
            </button>
          ))}
        </div>
        <div className="rounded-xl bg-pos-card border border-pos-border p-4">
          <div className="flex items-center gap-3 mb-4">
            <label className="text-sm text-pos-muted">Number of splits</label>
            <div className="flex items-center gap-1">
              <button
                className="w-9 h-9 rounded-lg bg-white/10 text-pos-text"
                onClick={() => setSplitCount(Math.max(2, seats - 1))}
              >
                −
              </button>
              <span className="w-9 text-center text-base font-bold text-pos-text">{seats}</span>
              <button
                className="w-9 h-9 rounded-lg bg-white/10 text-pos-text"
                onClick={() => setSplitCount(seats + 1)}
              >
                +
              </button>
            </div>
            <span className="text-xs text-pos-muted ml-auto">
              Even split: {formatCurrency(evenSplit)}
            </span>
          </div>

          {mode === 'seat' && (
            <div className="space-y-3">
              {items.map((it) => (
                <div
                  key={it.id}
                  className="flex items-center justify-between p-3 rounded-lg bg-white/6 border border-pos-border"
                >
                  <div>
                    <div className="text-sm text-pos-text">
                      {it.quantity}× {it.name}
                    </div>
                    <div className="text-xs text-pos-muted">
                      {formatCurrency(it.price * it.quantity)}
                    </div>
                  </div>
                  <select
                    value={assignments[it.id] ?? 0}
                    onChange={(e) =>
                      setAssignments((prev) => ({ ...prev, [it.id]: Number(e.target.value) }))
                    }
                    className="h-10 rounded-lg bg-white/10 text-pos-text px-3 text-sm"
                  >
                    {Array.from({ length: seats }).map((_, i) => (
                      <option key={i} value={i}>
                        Seat {i + 1}
                      </option>
                    ))}
                  </select>
                </div>
              ))}
              <div className="grid grid-cols-2 gap-2 mt-4">
                {seatTotals.map((t, i) => (
                  <div
                    key={i}
                    className="p-3 rounded-lg bg-white/6 border border-pos-border text-center"
                  >
                    <div className="text-xs text-pos-muted">Seat {i + 1}</div>
                    <div className="text-lg font-bold text-pos-text">{formatCurrency(t)}</div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {mode === 'amount' && (
            <div className="space-y-2">
              {amountSplits.map((raw, idx) => (
                <div key={idx} className="flex items-center gap-3">
                  <span className="text-sm text-pos-muted w-20">Split {idx + 1}</span>
                  <input
                    type="number"
                    inputMode="decimal"
                    value={raw}
                    onChange={(e) => updateAmount(idx, e.target.value)}
                    className="flex-1 h-11 px-3 rounded-lg bg-white/10 text-pos-text"
                  />
                </div>
              ))}
              <p
                className={`text-xs mt-2 ${
                  amountMatchesTotal ? 'text-pos-muted' : 'text-amber-300'
                }`}
              >
                Sum: {formatCurrency(amountSum)} / {formatCurrency(total)}
                {amountMatchesTotal ? '' : ' - must match the total'}
              </p>
            </div>
          )}

          {mode === 'items' && (
            <div className="text-sm text-pos-muted">
              Tap items in the Terminal to assign to a split; switch to "By seat" for full
              assignment.
            </div>
          )}
        </div>
        <button
          onClick={() => void confirm()}
          disabled={confirming || mode !== 'seat'}
          title={mode !== 'seat' ? 'Switch to By seat to confirm the split' : undefined}
          className="w-full h-14 rounded-xl bg-pos-text text-pos-bg font-bold text-base hover:opacity-90 transition disabled:opacity-40"
        >
          {confirming ? 'Splitting…' : 'Confirm Split'}
        </button>
      </div>
    </ScreenLayout>
  );
};

export default SplitBillScreen;
