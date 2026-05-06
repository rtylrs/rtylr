import { useEffect } from 'react';
import { useOfflineQueueStore } from '../../shared/stores/offlineQueueStore';
import { syncPendingOfflineOrders } from './offlineQueueSync';
import { syncPendingSessionCloses } from '../close-session/sessionCloseQueueSync';
import { syncPendingCashMovements } from '../cash-drop/cashMovementQueueSync';

export function useOfflineQueueAutoSync(): void {
  const online = useOfflineQueueStore((s) => s.online);

  useEffect(() => {
    if (!online) return;

    void syncPendingOfflineOrders();
    void syncPendingSessionCloses();
    void syncPendingCashMovements();

    const onReconnected = (): void => {
      void syncPendingOfflineOrders();
      void syncPendingSessionCloses();
      void syncPendingCashMovements();
    };

    window.addEventListener('rtylr-offline-queue-reconnected', onReconnected);
    return () => {
      window.removeEventListener('rtylr-offline-queue-reconnected', onReconnected);
    };
  }, [online]);
}
