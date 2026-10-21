import { createOneKitClient } from '@voxire/api';
import { SessionServiceClient } from '@voxire/voxie/pos/session/v1/client';
import React, { useMemo, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';
import ScreenLayout from '../../layouts/ScreenLayout';
import ConfirmModal from '../../shared/components/ConfirmModal';
import { formatCurrency } from '../../shared/utils/format';
import { useCashDrawerStore } from '../../shared/stores/cashDrawerStore';
import { useDaySummary } from '../day-summary/useDaySummary';
import { useSessionCloseQueueStore } from '../../shared/stores/sessionCloseQueueStore';
import { buildCloseSessionPayload, closeSessionWithFallback } from './closeSessionWorkflow';

const sessionClient = createOneKitClient(SessionServiceClient, 'pos');

const CloseSessionScreen: React.FC = () => {
  const navigate = useNavigate();
  const { summary } = useDaySummary();
  const movements = useCashDrawerStore((s) => s.movements);
  const openingFloat = useCashDrawerStore((s) => s.openingFloat);
  const expectedCashFn = useCashDrawerStore((s) => s.expectedCash);
  const expectedCash = useMemo(() => expectedCashFn(), [expectedCashFn, movements, openingFloat]);
  const closeDrawer = useCashDrawerStore((s) => s.closeDrawer);
  const enqueueSessionClose = useSessionCloseQueueStore((s) => s.enqueue);
  const [actualCash, setActualCash] = useState('');
  const [busy, setBusy] = useState(false);
  const [confirmZero, setConfirmZero] = useState(false);
  // Closing twice would open a second register session; a disabled button does
  // not cover two clicks inside one tick, so the latch flips before the await.
  const closingRef = useRef(false);

  let variance = 0;
  try {
    variance = buildCloseSessionPayload({
      expected: expectedCash,
      actualInput: actualCash || '0',
    }).variance;
  } catch {
    variance = -expectedCash;
  }

  const close = async () => {
    if (closingRef.current) return;
    closingRef.current = true;
    setBusy(true);
    try {
      const result = await closeSessionWithFallback({
        payload: buildCloseSessionPayload({
          expected: expectedCash,
          actualInput: actualCash || '0',
        }),
        closeSession: async (payload) => {
          await sessionClient.closeSession(payload);
        },
        enqueue: enqueueSessionClose,
      });
      closeDrawer();
      setConfirmZero(false);
      if (result.status === 'queued') {
        toast.warning('Session close saved locally and will sync later');
      } else {
        toast.success('Session closed');
      }
      navigate('/day-summary');
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Unable to close session.');
    } finally {
      // Released on failure too, so the cashier can retry from a still-open
      // dialog rather than being stuck on a permanently disabled till.
      closingRef.current = false;
      setBusy(false);
    }
  };

  // An empty input coerces to $0, so it must be an explicit confirmation —
  // never a silent default.
  const requestClose = () => {
    if (!actualCash.trim()) {
      setConfirmZero(true);
      return;
    }
    const parsed = Number(actualCash);
    if (!Number.isFinite(parsed) || parsed < 0) {
      toast.error('Enter a valid counted cash amount.');
      return;
    }
    if (parsed === 0) {
      setConfirmZero(true);
      return;
    }
    void close();
  };

  return (
    <ScreenLayout title="Close Session">
      <div className="p-5 max-w-xl space-y-5">
        <div className="rounded-xl bg-pos-card border border-pos-border p-4 space-y-2">
          <div className="flex justify-between text-sm">
            <span className="text-pos-muted">Expected cash in drawer</span>
            <span className="text-pos-text font-semibold">{formatCurrency(expectedCash)}</span>
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
          onClick={requestClose}
          disabled={busy}
          className="w-full h-14 rounded-xl bg-pos-text text-pos-bg font-bold disabled:opacity-50"
        >
          {busy ? 'Closing…' : 'Close Session'}
        </button>
      </div>
      <ConfirmModal
        open={confirmZero}
        title="Close with $0.00 counted cash?"
        message={`No counted cash was entered. This records a closing float of ${formatCurrency(0)} against an expected ${formatCurrency(expectedCash)}.`}
        confirmText={busy ? 'Closing…' : 'Close anyway'}
        pending={busy}
        danger
        onCancel={() => {
          if (closingRef.current) return;
          setConfirmZero(false);
        }}
        onConfirm={() => {
          // The dialog stays open (disabled) until the request settles, so a
          // failure leaves it in place to retry from.
          void close();
        }}
      />
    </ScreenLayout>
  );
};

export default CloseSessionScreen;
