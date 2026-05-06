import { renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { useOnlineStatus } from './useOnlineStatus';
import { useOfflineQueueStore } from '../stores/offlineQueueStore';

function setNavigatorOnline(value: boolean): void {
  Object.defineProperty(window.navigator, 'onLine', {
    configurable: true,
    value,
  });
}

describe('useOnlineStatus', () => {
  beforeEach(() => {
    useOfflineQueueStore.setState({ online: true, queue: [] });
    setNavigatorOnline(true);
  });

  it('syncs the initial browser online state when mounted', () => {
    setNavigatorOnline(false);

    renderHook(() => useOnlineStatus());

    expect(useOfflineQueueStore.getState().online).toBe(false);
  });
});
