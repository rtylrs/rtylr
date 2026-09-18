import React from 'react';
import { useLiveQuery } from 'dexie-react-hooks';
import { db } from '../services/db';
import { useOfflineQueueStore } from '../stores/offlineQueueStore';
import { useSessionCloseQueueStore } from '../stores/sessionCloseQueueStore';
import { useCashMovementQueueStore } from '../stores/cashMovementQueueStore';

const OfflineBadge: React.FC = () => {
  const online = useOfflineQueueStore((s) => s.online);
  const orderQueueSize = useOfflineQueueStore((s) => s.queue.length);
  const sessionCloseQueueSize = useSessionCloseQueueStore((s) => s.queue.length);
  const cashMovementQueueSize = useCashMovementQueueStore((s) => s.queue.length);
  const apiMutationQueueSize =
    useLiveQuery(
      () => db.apiMutations.where('status').anyOf('pending', 'syncing', 'failed').count(),
      [],
      0
    ) ?? 0;
  const queueSize =
    orderQueueSize + sessionCloseQueueSize + cashMovementQueueSize + apiMutationQueueSize;
  if (online && queueSize === 0) return null;
  return (
    <div
      className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-[11px] font-semibold ${
        online
          ? 'bg-amber-500/15 text-amber-300 border border-amber-500/30'
          : 'bg-red-500/15 text-red-300 border border-red-500/30'
      }`}
    >
      <span className={`w-2 h-2 rounded-full ${online ? 'bg-amber-400' : 'bg-red-400'}`} />
      {online ? `${queueSize} pending sync` : 'Offline'}
    </div>
  );
};

export default OfflineBadge;
