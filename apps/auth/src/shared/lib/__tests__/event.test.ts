import { describe, it, expect, vi } from 'vitest';
import { EventEmitter } from '../event';
type TestEvents = { login: { userId: string }; count: number };
describe('EventEmitter', () => {
  it('calls listeners when event is emitted', () => {
    const emitter = new EventEmitter<TestEvents>();
    const listener = vi.fn();
    emitter.on('login', listener);
    emitter.emit('login', { userId: 'u1' });
    expect(listener).toHaveBeenCalledWith({ userId: 'u1' });
  });
  it('unsubscribes via the returned function', () => {
    const emitter = new EventEmitter<TestEvents>();
    const listener = vi.fn();
    const unsub = emitter.on('count', listener);
    unsub();
    emitter.emit('count', 1);
    expect(listener).not.toHaveBeenCalled();
  });
  it('removes listener via off()', () => {
    const emitter = new EventEmitter<TestEvents>();
    const listener = vi.fn();
    emitter.on('count', listener);
    emitter.off('count', listener);
    emitter.emit('count', 1);
    expect(listener).not.toHaveBeenCalled();
  });
  it('clears all listeners', () => {
    const emitter = new EventEmitter<TestEvents>();
    const listener = vi.fn();
    emitter.on('count', listener);
    emitter.clear();
    emitter.emit('count', 1);
    expect(listener).not.toHaveBeenCalled();
  });
});
