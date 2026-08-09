import { describe, it, expect } from 'vitest';
import {
  ORDER_STATUS_STEPS,
  statusIndex,
  statusProgress,
  isOrderDone,
  statusMessage,
} from './orderStatusUtils';

describe('ORDER_STATUS_STEPS', () => {
  it('has four steps', () => {
    expect(ORDER_STATUS_STEPS).toHaveLength(4);
  });

  it('starts with received and ends with done', () => {
    expect(ORDER_STATUS_STEPS[0].key).toBe('received');
    expect(ORDER_STATUS_STEPS[ORDER_STATUS_STEPS.length - 1].key).toBe('done');
  });
});

describe('statusIndex', () => {
  it('returns 0 for received', () => {
    expect(statusIndex('received')).toBe(0);
  });

  it('returns 1 for preparing', () => {
    expect(statusIndex('preparing')).toBe(1);
  });

  it('returns 2 for ready', () => {
    expect(statusIndex('ready')).toBe(2);
  });

  it('returns 3 for done', () => {
    expect(statusIndex('done')).toBe(3);
  });
});

describe('statusProgress', () => {
  it('returns 25% for received (1 of 4)', () => {
    expect(statusProgress('received')).toBe(25);
  });

  it('returns 50% for preparing (2 of 4)', () => {
    expect(statusProgress('preparing')).toBe(50);
  });

  it('returns 75% for ready (3 of 4)', () => {
    expect(statusProgress('ready')).toBe(75);
  });

  it('returns 100% for done (4 of 4)', () => {
    expect(statusProgress('done')).toBe(100);
  });
});

describe('isOrderDone', () => {
  it('returns true when status is done', () => {
    expect(isOrderDone('done')).toBe(true);
  });

  it('returns false for received', () => {
    expect(isOrderDone('received')).toBe(false);
  });

  it('returns false for preparing', () => {
    expect(isOrderDone('preparing')).toBe(false);
  });

  it('returns false for ready', () => {
    expect(isOrderDone('ready')).toBe(false);
  });
});

describe('statusMessage', () => {
  it('returns the correct message for received', () => {
    expect(statusMessage('received')).toBe('Order received!');
  });

  it('returns the correct message for preparing', () => {
    expect(statusMessage('preparing')).toBe('Being prepared...');
  });

  it('returns the correct message for ready', () => {
    expect(statusMessage('ready')).toBe('Ready for pickup!');
  });

  it('returns the correct message for done', () => {
    expect(statusMessage('done')).toBe('Order complete!');
  });
});
