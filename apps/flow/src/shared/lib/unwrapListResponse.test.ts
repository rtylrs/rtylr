import { describe, it, expect } from 'vitest';

import { unwrapListResponse } from './unwrapListResponse';

describe('unwrapListResponse', () => {
  it('passes a raw array straight through', () => {
    const arr = [1, 2, 3];
    expect(unwrapListResponse(arr, 'tasks')).toEqual([1, 2, 3]);
  });

  it('unwraps a single-key envelope', () => {
    const envelope = { tasks: ['a', 'b'] };
    expect(unwrapListResponse(envelope, 'tasks')).toEqual(['a', 'b']);
  });

  it('returns empty array for a missing key in the envelope', () => {
    const envelope = { items: ['x'] };
    expect(unwrapListResponse(envelope, 'tasks')).toEqual([]);
  });

  it('returns empty array for null', () => {
    expect(unwrapListResponse(null, 'tasks')).toEqual([]);
  });

  it('returns empty array for undefined', () => {
    expect(unwrapListResponse(undefined, 'tasks')).toEqual([]);
  });

  it('returns empty array for a non-object primitive', () => {
    expect(unwrapListResponse(42, 'tasks')).toEqual([]);
    expect(unwrapListResponse('hello', 'tasks')).toEqual([]);
  });

  it('returns empty array when the keyed value is not an array', () => {
    expect(unwrapListResponse({ tasks: 'not-an-array' }, 'tasks')).toEqual([]);
    expect(unwrapListResponse({ tasks: null }, 'tasks')).toEqual([]);
  });

  it('uses the correct key - sibling keys do not bleed through', () => {
    const envelope = { projects: ['p1'], tasks: ['t1', 't2'] };
    expect(unwrapListResponse(envelope, 'projects')).toEqual(['p1']);
    expect(unwrapListResponse(envelope, 'tasks')).toEqual(['t1', 't2']);
  });
});
