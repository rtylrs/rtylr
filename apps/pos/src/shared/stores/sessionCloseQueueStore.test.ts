import { beforeEach, describe, expect, it } from 'vitest';
import { useSessionCloseQueueStore } from './sessionCloseQueueStore';

const { getState, setState } = useSessionCloseQueueStore;

beforeEach(() => {
  setState({ queue: [] });
});

describe('sessionCloseQueueStore', () => {
  it('enqueues a close-session payload for later sync', () => {
    const queued = getState().enqueue({
      sessionId: 'current',
      expected: 100,
      actual: 95,
      variance: -5,
    });

    expect(queued).toMatchObject({
      payload: {
        sessionId: 'current',
        expected: 100,
        actual: 95,
        variance: -5,
      },
      retries: 0,
      status: 'pending',
    });
    expect(getState().queue).toHaveLength(1);
  });
});
