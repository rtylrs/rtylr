import { describe, it, expect } from 'vitest';
import {
  pluralize,
  countLabel,
  truncate,
  capitalize,
  toLabel,
  ordinal,
  joinList,
  isBlank,
} from './labelUtils';

describe('pluralize', () => {
  it('singular for 1', () => expect(pluralize(1, 'item')).toBe('item'));
  it('plural for 0', () => expect(pluralize(0, 'item')).toBe('items'));
  it('plural for 2', () => expect(pluralize(2, 'item')).toBe('items'));
  it('custom plural', () => expect(pluralize(2, 'ox', 'oxen')).toBe('oxen'));
});

describe('countLabel', () => {
  it('1 item', () => expect(countLabel(1, 'item')).toBe('1 item'));
  it('3 items', () => expect(countLabel(3, 'item')).toBe('3 items'));
});

describe('truncate', () => {
  it('no-op when short enough', () => expect(truncate('hello', 10)).toBe('hello'));
  it('truncates long string', () => expect(truncate('hello world', 8)).toBe('hello...'));
  it('custom ellipsis', () => expect(truncate('hello world', 7, '~')).toBe('hello ~'));
});

describe('capitalize', () => {
  it('capitalizes first char', () => expect(capitalize('hello')).toBe('Hello'));
  it('empty string', () => expect(capitalize('')).toBe(''));
});

describe('toLabel', () => {
  it('camelCase', () => expect(toLabel('firstName')).toBe('First Name'));
  it('snake_case', () => expect(toLabel('first_name')).toBe('First Name'));
  it('already spaced', () => expect(toLabel('name')).toBe('Name'));
});

describe('ordinal', () => {
  it('1st', () => expect(ordinal(1)).toBe('1st'));
  it('2nd', () => expect(ordinal(2)).toBe('2nd'));
  it('3rd', () => expect(ordinal(3)).toBe('3rd'));
  it('4th', () => expect(ordinal(4)).toBe('4th'));
  it('11th', () => expect(ordinal(11)).toBe('11th'));
  it('21st', () => expect(ordinal(21)).toBe('21st'));
  it('12th', () => expect(ordinal(12)).toBe('12th'));
});

describe('joinList', () => {
  it('empty', () => expect(joinList([])).toBe(''));
  it('one item', () => expect(joinList(['a'])).toBe('a'));
  it('two items', () => expect(joinList(['a', 'b'])).toBe('a and b'));
  it('three items', () => expect(joinList(['a', 'b', 'c'])).toBe('a, b, and c'));
  it('custom conjunction', () => expect(joinList(['a', 'b'], 'or')).toBe('a or b'));
});

describe('isBlank', () => {
  it('empty string', () => expect(isBlank('')).toBe(true));
  it('whitespace', () => expect(isBlank('   ')).toBe(true));
  it('null', () => expect(isBlank(null)).toBe(true));
  it('undefined', () => expect(isBlank(undefined)).toBe(true));
  it('non-blank', () => expect(isBlank('hello')).toBe(false));
});
