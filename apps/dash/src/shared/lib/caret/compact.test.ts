import { describe, expect, it } from 'vitest';
import { compact } from './compact';

describe('compact', () => {
  it('drops null and undefined', () => {
    expect(compact([1, null, 2, undefined, 3])).toEqual([1, 2, 3]);
  });
  it('keeps falsy non-nullish values', () => {
    expect(compact([0, '', false])).toEqual([0, '', false]);
  });
});
