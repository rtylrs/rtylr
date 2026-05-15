import { describe, expect, it } from 'vitest';
import { scoreChat } from './searchScore';

describe('scoreChat', () => {
  it('returns 0 for empty query', () => {
    expect(scoreChat({ title: 'hi' }, '')).toBe(0);
  });

  it('scores exact title matches highest', () => {
    expect(scoreChat({ title: 'orders' }, 'orders')).toBeGreaterThan(
      scoreChat({ title: 'all orders today' }, 'orders')
    );
  });

  it('boosts title prefix matches', () => {
    expect(scoreChat({ title: 'orders today' }, 'orders')).toBeGreaterThan(
      scoreChat({ title: 'today orders' }, 'orders')
    );
  });

  it('considers the summary field', () => {
    expect(scoreChat({ title: 'foo', summary: 'orders here' }, 'orders')).toBeGreaterThan(0);
  });
});
