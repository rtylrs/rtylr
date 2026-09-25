import { beforeEach, describe, expect, it } from 'vitest';
import { useSessionCloseQueueStore } from './sessionCloseQueueStore';

const { getState, setState } = useSessionCloseQueueStore;

beforeEach(() => {
  setState({ queue: [] });
});

describe('sessionCloseQueueStore', () => {
  it('enqueues a close-session payload for later sync', () => {
    const queued = getState().enqueue({
      closingFloat: 95,
      expected: 100,
      variance: -5,
      sessionId: 12,
    });

    expect(queued).toMatchObject({
      payload: {
        closingFloat: 95,
        expected: 100,
        variance: -5,
        sessionId: 12,
      },
      retries: 0,
      status: 'pending',
    });
    expect(getState().queue).toHaveLength(1);
  });
});
