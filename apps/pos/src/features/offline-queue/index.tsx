import React, { useState } from 'react';
import { WifiOff } from 'lucide-react';
import { toast } from 'sonner';
import ScreenLayout from '../../layouts/ScreenLayout';
import EmptyState from '../../shared/components/EmptyState';
import ConfirmModal from '../../shared/components/ConfirmModal';
import { useOfflineQueue } from './useOfflineQueue';
import { formatCurrency, formatDate } from '../../shared/utils/format';

const OfflineQueueScreen: React.FC = () => {
  const {
    queue,
    sessionCloseQueue,
    cashMovementQueue,
    totalQueued,
    online,
    retry,
    retrySessionClose,
    retryCashMovement,
    flush,
    clear,
  } = useOfflineQueue();

  const [confirmingClear, setConfirmingClear] = useState(false);
  const [flushing, setFlushing] = useState(false);

  const doFlush = async () => {
    if (flushing) return;
    setFlushing(true);
    try {
      const result = await flush();
      const failed = result.failed;
      if (failed > 0) {
        toast.warning(`Sync attempted: ${result.synced} sent, ${failed} failed. Check the entries below.`);
      } else {
        toast.success('Queue synced');
      }
    } catch (error) {
      toast.error(
        error instanceof Error ? `Sync failed: ${error.message}` : 'Sync failed. Try again.',
      );
    } finally {
      setFlushing(false);
    }
  };

  const retryWithFeedback = async (id: string, retryFn: (id: string) => Promise<boolean>) => {
    try {
      if (await retryFn(id)) {
        toast.success('Entry synced');
      } else {
        toast.error('Retry failed — the entry stays queued');
      }
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Retry failed — the entry stays queued');
    }
  };

  return (
    <ScreenLayout
      title="Offline Queue"
      subtitle={online ? 'Online' : 'Offline'}
      actions={
        <div className="flex gap-2">
          <button
            onClick={doFlush}
            disabled={!online || totalQueued === 0}
            className="h-10 px-4 rounded-lg bg-pos-text text-pos-bg font-semibold text-sm disabled:opacity-50"
          >
            Sync All
          </button>
          <button
            onClick={() => setConfirmingClear(true)}
            disabled={totalQueued === 0}
            className="h-10 px-4 rounded-lg bg-destructive text-white font-semibold text-sm disabled:opacity-50"
          >
            Clear
          </button>
        </div>
      }
    >
      <div className="p-5 max-w-4xl">
        {totalQueued === 0 ? (
          <EmptyState
            icon={WifiOff}
            title="Nothing queued"
            message="Offline orders, session closes, and cash movements will appear here."
          />
        ) : (
          <ul className="space-y-2">
            {queue.map((q) => (
              <li
                key={q.id}
                className="p-4 rounded-xl bg-pos-card border border-pos-border flex items-center justify-between"
              >
                <div>
                  <div className="text-sm font-semibold text-pos-text">
                    Order · {(q.order.items ?? []).length} item
                    {(q.order.items ?? []).length === 1 ? '' : 's'}
                  </div>
                  <div className="text-xs text-pos-muted">
                    Queued {formatDate(q.queuedAt)} · Retries {q.retries} · {q.status}
                  </div>
                  {q.lastError && (
                    <div className="text-xs text-destructive mt-1">{q.lastError}</div>
                  )}
                </div>
                <div className="flex items-center gap-3">
                  <div className="text-sm font-bold text-pos-text">
                    {formatCurrency(
                      (q.order.items ?? []).reduce(
                        (sum, it) => sum + (it.unitPrice ?? 0) * (it.quantity ?? 0),
                        0
                      )
                    )}
                  </div>
                  <button
                    onClick={() => void retryWithFeedback(q.id, retry)}
                    disabled={!online}
                    className="h-10 px-3 rounded-lg bg-white/10 text-pos-text text-xs font-semibold disabled:opacity-50"
                  >
                    Retry
                  </button>
                </div>
              </li>
            ))}
            {sessionCloseQueue.map((q) => (
              <li
                key={q.id}
                className="p-4 rounded-xl bg-pos-card border border-pos-border flex items-center justify-between"
              >
                <div>
                  <div className="text-sm font-semibold text-pos-text">Session close</div>
                  <div className="text-xs text-pos-muted">
                    Queued {formatDate(q.queuedAt)} · Retries {q.retries} · {q.status}
                  </div>
                  {q.lastError && (
                    <div className="text-xs text-destructive mt-1">{q.lastError}</div>
                  )}
                </div>
                <div className="flex items-center gap-3">
                  <div className="text-sm font-bold text-pos-text">
                    Variance {formatCurrency(Number(q.payload.variance ?? 0))}
                  </div>
                  <button
                    onClick={() => void retryWithFeedback(q.id, retrySessionClose)}
                    disabled={!online}
                    className="h-10 px-3 rounded-lg bg-white/10 text-pos-text text-xs font-semibold disabled:opacity-50"
                  >
                    Retry
                  </button>
                </div>
              </li>
            ))}
            {cashMovementQueue.map((q) => (
              <li
                key={q.id}
                className="p-4 rounded-xl bg-pos-card border border-pos-border flex items-center justify-between"
              >
                <div>
                  <div className="text-sm font-semibold text-pos-text">
                    Cash movement · {q.action.replace('_', ' ')}
                  </div>
                  <div className="text-xs text-pos-muted">
                    Queued {formatDate(q.queuedAt)} · Retries {q.retries} · {q.status}
                  </div>
                  {q.payload.reason && (
                    <div className="text-xs text-pos-muted mt-1">{q.payload.reason}</div>
                  )}
                  {q.lastError && (
                    <div className="text-xs text-destructive mt-1">{q.lastError}</div>
                  )}
                </div>
                <div className="flex items-center gap-3">
                  <div className="text-sm font-bold text-pos-text">
                    {formatCurrency(q.payload.amount ?? 0)}
                  </div>
                  <button
                    onClick={() => void retryWithFeedback(q.id, retryCashMovement)}
                    disabled={!online}
                    className="h-10 px-3 rounded-lg bg-white/10 text-pos-text text-xs font-semibold disabled:opacity-50"
                  >
                    Retry
                  </button>
                </div>
              </li>
            ))}
          </ul>
        )}
      </div>

      <ConfirmModal
        open={confirmingClear}
        title="Clear the offline queue?"
        message="Unsynced orders, session closes, and cash movements will be permanently discarded."
        danger
        confirmText="Discard"
        onConfirm={() => {
          clear();
          setConfirmingClear(false);
        }}
        onCancel={() => setConfirmingClear(false)}
      />
    </ScreenLayout>
  );
};

export default OfflineQueueScreen;
