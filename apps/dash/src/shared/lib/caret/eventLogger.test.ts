import { describe, expect, it } from 'vitest';
import { logEvent, subscribeEvents, type CaretEvent } from './eventLogger';

describe('eventLogger', () => {
  it('delivers events to subscribers', () => {
    const events: CaretEvent[] = [];
    const off = subscribeEvents((e) => events.push(e));
    logEvent('test', { foo: 'bar' });
    off();
    expect(events).toHaveLength(1);
    expect(events[0].name).toBe('test');
    expect(events[0].props?.foo).toBe('bar');
  });

  it('isolates listener errors', () => {
    const off1 = subscribeEvents(() => {
      throw new Error('bad');
    });
    let received = 0;
    const off2 = subscribeEvents(() => {
      received += 1;
    });
    logEvent('x');
    off1();
    off2();
    expect(received).toBe(1);
  });
});
