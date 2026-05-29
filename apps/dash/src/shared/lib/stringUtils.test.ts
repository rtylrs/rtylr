import { describe, it, expect } from 'vitest';
import {
  truncate,
  capitalize,
  camelToLabel,
  snakeToTitle,
  kebabToTitle,
  padStart,
  stripWhitespace,
  countOccurrences,
} from './stringUtils';

describe('truncate', () => {
  it('returns the original string when within limit', () => {
    expect(truncate('hello', 10)).toBe('hello');
  });

  it('truncates and appends ellipsis', () => {
    expect(truncate('hello world', 8)).toBe('hello...');
  });

  it('handles empty string', () => {
    expect(truncate('', 5)).toBe('');
  });
});

describe('capitalize', () => {
  it('capitalizes first letter', () => {
    expect(capitalize('hello')).toBe('Hello');
  });

  it('returns empty string unchanged', () => {
    expect(capitalize('')).toBe('');
  });
});

describe('camelToLabel', () => {
  it('converts camelCase to label', () => {
    expect(camelToLabel('myFieldName')).toBe('My Field Name');
  });

  it('handles single word', () => {
    expect(camelToLabel('name')).toBe('Name');
  });
});

describe('snakeToTitle', () => {
  it('converts snake_case to Title Case', () => {
    expect(snakeToTitle('my_field_name')).toBe('My Field Name');
  });

  it('handles single word', () => {
    expect(snakeToTitle('name')).toBe('Name');
  });
});

describe('kebabToTitle', () => {
  it('converts kebab-case to Title Case', () => {
    expect(kebabToTitle('my-field-name')).toBe('My Field Name');
  });

  it('handles single word', () => {
    expect(kebabToTitle('name')).toBe('Name');
  });
});

describe('padStart', () => {
  it('pads shorter string to given width', () => {
    expect(padStart('5', 3, '0')).toBe('005');
  });

  it('returns string unchanged if already at width', () => {
    expect(padStart('abc', 3)).toBe('abc');
  });
});

describe('stripWhitespace', () => {
  it('removes all whitespace', () => {
    expect(stripWhitespace('h e l l o')).toBe('hello');
  });

  it('returns empty string unchanged', () => {
    expect(stripWhitespace('')).toBe('');
  });
});

describe('countOccurrences', () => {
  it('counts non-overlapping occurrences', () => {
    expect(countOccurrences('banana', 'an')).toBe(2);
  });

  it('returns 0 when substring not found', () => {
    expect(countOccurrences('hello', 'xyz')).toBe(0);
  });

  it('returns 0 for empty substring', () => {
    expect(countOccurrences('hello', '')).toBe(0);
  });
});
