import { describe, expect, it } from 'vitest';
import { inferIntent } from './intent';

describe('inferIntent', () => {
  it('flags write verbs', () => {
    expect(inferIntent('Create a customer named Bob')).toBe('write');
    expect(inferIntent('Refund order 42')).toBe('write');
    expect(inferIntent('Remove that user')).toBe('write');
  });

  it('flags read phrasing', () => {
    expect(inferIntent('Show me the orders')).toBe('read');
    expect(inferIntent('How many staff are clocked in?')).toBe('read');
    expect(inferIntent("What's my revenue?")).toBe('read');
  });

  it('returns unknown when no keywords match', () => {
    expect(inferIntent('hi there')).toBe('unknown');
  });
});
