import { beforeEach, describe, expect, it, vi } from 'vitest';
import {
  retrySessionClose,
  syncAllSessionCloses,
  syncPendingSessionCloses,
} from './sessionCloseQueueSync';
import { useSessionCloseQueueStore } from '../../shared/stores/sessionCloseQueueStore';

beforeEach(() => {
  useSessionCloseQueueStore.setState({ queue: [] });
});

describe('syncPendingSessionCloses', () => {
  it('syncs only pending queued session closes from the store', async () => {
    const first = useSessionCloseQueueStore.getState().enqueue({
      sessionId: 42,
      closingFloat: 95,
      expected: 100,
      variance: -5,
    });
    const failed = useSessionCloseQueueStore.getState().enqueue({
      sessionId: 87,
      closingFloat: 200,
      expected: 200,
      variance: 0,
    });
    useSessionCloseQueueStore.getState().markFailed(failed.id, 'previous failure');

    const syncClose = vi.fn(async () => undefined);

    const result = await syncPendingSessionCloses(syncClose);

    expect(result).toEqual({ attempted: 1, synced: 1, failed: 0 });
    expect(syncClose).toHaveBeenCalledOnce();
    expect(syncClose).toHaveBeenCalledWith(first.payload);
    expect(useSessionCloseQueueStore.getState().queue).toHaveLength(1);
    expect(useSessionCloseQueueStore.getState().queue[0].id).toBe(failed.id);
  });
});

describe('syncAllSessionCloses', () => {
  it('syncs failed and pending queued session closes from the store', async () => {
    const pending = useSessionCloseQueueStore.getState().enqueue({
      sessionId: 42,
      closingFloat: 95,
      expected: 100,
      variance: -5,
    });
    const failed = useSessionCloseQueueStore.getState().enqueue({
      sessionId: 87,
      closingFloat: 200,
      expected: 200,
      variance: 0,
    });
    useSessionCloseQueueStore.getState().markFailed(failed.id, 'previous failure');

    const syncClose = vi.fn(async () => undefined);

    const result = await syncAllSessionCloses(syncClose);

    expect(result).toEqual({ attempted: 2, synced: 2, failed: 0 });
    expect(syncClose).toHaveBeenCalledWith(pending.payload);
    expect(syncClose).toHaveBeenCalledWith(failed.payload);
    expect(useSessionCloseQueueStore.getState().queue).toHaveLength(0);
  });
});

describe('retrySessionClose', () => {
  it('retries a single queued session close by id', async () => {
    const failed = useSessionCloseQueueStore.getState().enqueue({
      sessionId: 42,
      closingFloat: 95,
      expected: 100,
      variance: -5,
    });
    useSessionCloseQueueStore.getState().markFailed(failed.id, 'previous failure');
    const other = useSessionCloseQueueStore.getState().enqueue({
      sessionId: 87,
      closingFloat: 200,
      expected: 200,
      variance: 0,
    });

    const syncClose = vi.fn(async () => undefined);

    const result = await retrySessionClose(failed.id, syncClose);

    expect(result).toBe(true);
    expect(syncClose).toHaveBeenCalledWith(failed.payload);
    expect(useSessionCloseQueueStore.getState().queue).toHaveLength(1);
    expect(useSessionCloseQueueStore.getState().queue[0].id).toBe(other.id);
  });

  it('returns false for an unknown session close queue id', async () => {
    const syncClose = vi.fn(async () => undefined);

    await expect(retrySessionClose('missing', syncClose)).resolves.toBe(false);

    expect(syncClose).not.toHaveBeenCalled();
  });
});
