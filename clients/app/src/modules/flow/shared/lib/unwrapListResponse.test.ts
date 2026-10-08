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

  it('rejects an envelope with a missing key', () => {
    const envelope = { items: ['x'] };
    expect(() => unwrapListResponse(envelope, 'tasks')).toThrow(
      'Expected a list response with "tasks"'
    );
  });

  it('rejects null, missing, and primitive responses', () => {
    for (const response of [null, undefined, 42, 'hello']) {
      expect(() => unwrapListResponse(response, 'tasks')).toThrow(
        'Expected a list response with "tasks"'
      );
    }
  });

  it('rejects a candidate key whose value is not an array', () => {
    expect(() => unwrapListResponse({ tasks: 'not-an-array' }, 'tasks')).toThrow(
      'Expected a list response with "tasks"'
    );
    expect(() => unwrapListResponse({ tasks: null }, 'tasks')).toThrow(
      'Expected a list response with "tasks"'
    );
  });

  it('uses the correct key - sibling keys do not bleed through', () => {
    const envelope = { projects: ['p1'], tasks: ['t1', 't2'] };
    expect(unwrapListResponse(envelope, 'projects')).toEqual(['p1']);
    expect(unwrapListResponse(envelope, 'tasks')).toEqual(['t1', 't2']);
  });
});
