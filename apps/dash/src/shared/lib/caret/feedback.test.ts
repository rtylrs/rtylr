import { describe, expect, it, beforeEach } from 'vitest';
import { recordVote, drainVotes } from './feedback';

beforeEach(() => {
  drainVotes(); // reset between tests
});

describe('feedback buffer', () => {
  it('records and drains a vote', () => {
    recordVote('m1', 'up');
    const out = drainVotes();
    expect(out).toHaveLength(1);
    expect(out[0].messageId).toBe('m1');
  });

  it('drains returns nothing on second call', () => {
    recordVote('m1', 'up');
    drainVotes();
    expect(drainVotes()).toEqual([]);
  });

  it('truncates beyond MAX', () => {
    for (let i = 0; i < 100; i++) recordVote('m' + i, 'up');
    expect(drainVotes().length).toBeLessThanOrEqual(50);
  });
});
