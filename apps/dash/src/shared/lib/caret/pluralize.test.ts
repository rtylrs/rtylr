import { describe, expect, it } from 'vitest';
import { pluralize } from './pluralize';

describe('pluralize', () => {
  it('uses singular for 1', () => {
    expect(pluralize(1, 'order')).toBe('1 order');
  });
  it('uses regular plural for other counts', () => {
    expect(pluralize(3, 'order')).toBe('3 orders');
  });
  it('accepts an irregular plural', () => {
    expect(pluralize(2, 'person', 'people')).toBe('2 people');
  });
  it('formats large numbers with separators', () => {
    expect(pluralize(1234, 'item')).toBe('1,234 items');
  });
});
