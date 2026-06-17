import React, { useMemo, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { toast } from 'sonner';
import { unwrap } from '@voxire/api';
import ScreenLayout from '../../layouts/ScreenLayout';
import { formatCurrency } from '../../shared/utils/format';
import { useCartStore } from '../../shared/stores/cartStore';
import { posService } from '../../shared/services/api/posService';
import { buildSplitBillRequest } from './splitBillWorkflow';

type SplitMode = 'seat' | 'amount' | 'items';

const SplitBillScreen: React.FC = () => {
  const { orderId } = useParams<{ orderId: string }>();
  const navigate = useNavigate();
  const items = useCartStore((s) => s.items);
  const total = useCartStore((s) => s.total());
  const serverOrderId = useCartStore((s) => s.serverOrderId);
  const [mode, setMode] = useState<SplitMode>('seat');
  const [seats, setSeats] = useState(2);
  const [assignments, setAssignments] = useState<Record<string, number>>({});
  const [amountSplits, setAmountSplits] = useState<number[]>([total / 2, total / 2]);
  const [confirming, setConfirming] = useState(false);

  const seatTotals = useMemo(() => {
    const result: number[] = Array.from({ length: seats }, () => 0);
    items.forEach((item) => {
      const seat = assignments[item.id] ?? 0;
      result[seat] = (result[seat] ?? 0) + item.price * item.quantity;
    });
    return result;
  }, [items, assignments, seats]);

  const evenSplit = total / seats;

  const setSplitCount = (n: number) => {
    setSeats(n);
    setAmountSplits(Array.from({ length: n }, () => total / n));
  };

  const updateAmount = (idx: number, value: number) => {
    setAmountSplits((prev) => prev.map((v, i) => (i === idx ? value : v)));
  };

  const confirm = async (): Promise<void> => {
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
      toast.success(`Split into ${seats} ticket${seats === 1 ? '' : 's'} (local)`);
      navigate(-1);
      return;
    }
    setConfirming(true);
    try {
      await unwrap(posService.splitOrder(request.orderId, request.payload));
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
              {amountSplits.map((amt, idx) => (
                <div key={idx} className="flex items-center gap-3">
                  <span className="text-sm text-pos-muted w-20">Split {idx + 1}</span>
                  <input
                    type="number"
                    value={amt.toFixed(2)}
                    onChange={(e) => updateAmount(idx, Number(e.target.value))}
                    className="flex-1 h-11 px-3 rounded-lg bg-white/10 text-pos-text"
                  />
                </div>
              ))}
              <p className="text-xs text-pos-muted mt-2">
                Sum: {formatCurrency(amountSplits.reduce((a, b) => a + b, 0))} /{' '}
                {formatCurrency(total)}
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
          onClick={confirm}
          disabled={confirming}
          className="w-full h-14 rounded-xl bg-pos-text text-pos-bg font-bold text-base hover:opacity-90 transition disabled:opacity-40"
        >
          {confirming ? 'Splitting…' : 'Confirm Split'}
        </button>
      </div>
    </ScreenLayout>
  );
};

export default SplitBillScreen;
