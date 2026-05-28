import { describe, it, expect } from 'vitest';
import { isPosEvent, KDS_EVENT_TYPES } from './realtimeEvents';

describe('isPosEvent', () => {
  it('returns true when the event type matches', () => {
    const event = { type: 'pos:order:paid', payload: {} };
    expect(isPosEvent(event, 'pos:order:paid')).toBe(true);
  });

  it('returns false when the event type does not match', () => {
    const event = { type: 'pos:order:created', payload: {} };
    expect(isPosEvent(event, 'pos:order:paid')).toBe(false);
  });

  it('returns false for an unrelated event type', () => {
    const event = { type: 'some:other:event' };
    expect(isPosEvent(event, 'pos:order:created')).toBe(false);
  });

  it('acts as a type guard (narrows the type)', () => {
    const events: { type: string; payload?: unknown }[] = [
      { type: 'pos:order:created', payload: { orderId: 1, orderNumber: 'ORD-1', locationId: 2, tableId: null, status: 'open', total: 20, createdAt: '2024-01-01T00:00:00Z' } },
      { type: 'pos:order:paid', payload: {} },
    ];
    const created = events.filter((e) => isPosEvent(e, 'pos:order:created'));
    expect(created).toHaveLength(1);
  });
});

describe('KDS_EVENT_TYPES', () => {
  it('contains pos:order:created', () => {
    expect(KDS_EVENT_TYPES.has('pos:order:created')).toBe(true);
  });

  it('contains pos:order:preparing', () => {
    expect(KDS_EVENT_TYPES.has('pos:order:preparing')).toBe(true);
  });

  it('contains pos:kds:item:updated', () => {
    expect(KDS_EVENT_TYPES.has('pos:kds:item:updated')).toBe(true);
  });

  it('does not contain non-KDS events', () => {
    expect(KDS_EVENT_TYPES.has('pos:order:paid')).toBe(false);
    expect(KDS_EVENT_TYPES.has('pos:table:status')).toBe(false);
    expect(KDS_EVENT_TYPES.has('pos:stock:low')).toBe(false);
  });

  it('has exactly 3 members', () => {
    expect(KDS_EVENT_TYPES.size).toBe(3);
  });
});
