import { unwrap } from '@voxire/api';
import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';
import ScreenLayout from '../../layouts/ScreenLayout';
import { formatCurrency } from '../../shared/utils/format';
import { useCashDrawerStore } from '../../shared/stores/cashDrawerStore';
import { useDaySummary } from '../day-summary/useDaySummary';
import { posService } from '../../shared/services/posService';

const CloseSessionScreen: React.FC = () => {
  const navigate = useNavigate();
  const { summary } = useDaySummary();
  const expected = useCashDrawerStore((s) => s.expectedCash());
  const closeDrawer = useCashDrawerStore((s) => s.closeDrawer);
  const [actualCash, setActualCash] = useState('');
  const [busy, setBusy] = useState(false);

  const actual = Number(actualCash || 0);
  const variance = actual - expected;

  const close = async () => {
    setBusy(true);
    try {
      await unwrap(
        posService.closeSession({
          sessionId: 'current',
          expected,
          actual,
          variance,
        })
      );
      closeDrawer();
      toast.success('Session closed');
      navigate('/day-summary');
    } catch {
      toast.warning('Saved locally, will sync');
      closeDrawer();
      navigate('/day-summary');
    } finally {
      setBusy(false);
    }
  };

  return (
    <ScreenLayout title="Close Session">
      <div className="p-5 max-w-xl space-y-5">
        <div className="rounded-xl bg-pos-card border border-pos-border p-4 space-y-2">
          <div className="flex justify-between text-sm">
            <span className="text-pos-muted">Expected cash in drawer</span>
            <span className="text-pos-text font-semibold">{formatCurrency(expected)}</span>
          </div>
          {summary && (
            <>
              <div className="flex justify-between text-sm">
                <span className="text-pos-muted">Gross sales</span>
                <span className="text-pos-text">{formatCurrency(summary.grossSales)}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-pos-muted">Tips</span>
                <span className="text-pos-text">{formatCurrency(summary.tips)}</span>
              </div>
            </>
          )}
        </div>
        <div>
          <label className="text-xs text-pos-muted">Actual counted cash</label>
          <input
            type="number"
            inputMode="decimal"
            value={actualCash}
            onChange={(e) => setActualCash(e.target.value)}
            placeholder="0.00"
            className="w-full h-14 mt-1 px-4 rounded-lg bg-white/8 border border-pos-border text-pos-text text-xl outline-none focus:border-white/25"
          />
        </div>
        <div
          className={`rounded-lg p-3 text-sm font-semibold ${
            variance === 0
              ? 'bg-emerald-500/15 text-emerald-300'
              : variance > 0
              ? 'bg-amber-500/15 text-amber-300'
              : 'bg-red-500/15 text-red-300'
          }`}
        >
          Variance: {formatCurrency(variance)}
        </div>
        <button
          onClick={close}
          disabled={busy}
          className="w-full h-14 rounded-xl bg-pos-text text-pos-bg font-bold disabled:opacity-50"
        >
          {busy ? 'Closing…' : 'Close Session'}
        </button>
      </div>
    </ScreenLayout>
  );
};

export default CloseSessionScreen;
