import { useEffect } from 'react';
import { useOfflineQueueStore } from '../stores/offlineQueueStore';

export function useOnlineStatus(): void {
  const setOnline = useOfflineQueueStore((s) => s.setOnline);
  useEffect(() => {
    const onOnline = (): void => setOnline(true);
    const onOffline = (): void => setOnline(false);
    setOnline(navigator.onLine);
    window.addEventListener('online', onOnline);
    window.addEventListener('offline', onOffline);
    return () => {
      window.removeEventListener('online', onOnline);
      window.removeEventListener('offline', onOffline);
    };
  }, [setOnline]);
}
