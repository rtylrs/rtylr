import { describe, it, expect } from 'vitest';

import { hasQueryFailed } from './queryFailureUtils';

describe('hasQueryFailed', () => {
  it('is true once the query has settled into a real error', () => {
    expect(hasQueryFailed({ isError: true, isPaused: false, failureCount: 1 })).toBe(true);
  });

  it('is true when the query is paused after already failing at least once', () => {
    expect(hasQueryFailed({ isError: false, isPaused: true, failureCount: 1 })).toBe(true);
  });

  it('is false when the query is paused before its very first attempt', () => {
    expect(hasQueryFailed({ isError: false, isPaused: true, failureCount: 0 })).toBe(false);
  });

  it('is false for a query that is simply still loading', () => {
    expect(hasQueryFailed({ isError: false, isPaused: false, failureCount: 0 })).toBe(false);
  });

  it('is false for a query that resolved successfully', () => {
    expect(hasQueryFailed({ isError: false, isPaused: false, failureCount: 0 })).toBe(false);
  });
});
