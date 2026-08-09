import { describe, it, expect } from 'vitest';
import {
  capitalise,
  toKebabCase,
  toTitleCase,
  isBlank,
  stripHtml,
  wordCount,
  repeat,
  padLeft,
} from './stringUtils';

describe('capitalise', () => {
  it('capitalises the first letter', () => {
    expect(capitalise('hello')).toBe('Hello');
  });

  it('leaves an already-capitalised string unchanged', () => {
    expect(capitalise('Hello')).toBe('Hello');
  });

  it('returns an empty string unchanged', () => {
    expect(capitalise('')).toBe('');
  });
});

describe('toKebabCase', () => {
  it('converts camelCase to kebab-case', () => {
    expect(toKebabCase('myVariableName')).toBe('my-variable-name');
  });

  it('converts PascalCase to kebab-case', () => {
    expect(toKebabCase('MyComponent')).toBe('my-component');
  });

  it('converts spaces to hyphens', () => {
    expect(toKebabCase('hello world')).toBe('hello-world');
  });
});

describe('toTitleCase', () => {
  it('converts kebab-case to Title Case', () => {
    expect(toTitleCase('hello-world')).toBe('Hello World');
  });

  it('converts snake_case to Title Case', () => {
    expect(toTitleCase('my_item_name')).toBe('My Item Name');
  });
});

describe('isBlank', () => {
  it('returns true for an empty string', () => {
    expect(isBlank('')).toBe(true);
  });

  it('returns true for whitespace only', () => {
    expect(isBlank('   ')).toBe(true);
  });

  it('returns false for a non-blank string', () => {
    expect(isBlank('hello')).toBe(false);
  });
});

describe('stripHtml', () => {
  it('removes HTML tags', () => {
    expect(stripHtml('<p>Hello <b>World</b></p>')).toBe('Hello World');
  });

  it('leaves plain text unchanged', () => {
    expect(stripHtml('Hello World')).toBe('Hello World');
  });
});

describe('wordCount', () => {
  it('counts words in a sentence', () => {
    expect(wordCount('hello world foo')).toBe(3);
  });

  it('returns 0 for an empty string', () => {
    expect(wordCount('')).toBe(0);
  });

  it('collapses multiple spaces', () => {
    expect(wordCount('a   b')).toBe(2);
  });
});

describe('repeat', () => {
  it('repeats a string n times', () => {
    expect(repeat('ab', 3)).toBe('ababab');
  });

  it('returns an empty string for 0 times', () => {
    expect(repeat('ab', 0)).toBe('');
  });

  it('returns an empty string for negative times', () => {
    expect(repeat('ab', -1)).toBe('');
  });
});

describe('padLeft', () => {
  it('pads with spaces by default', () => {
    expect(padLeft('5', 3)).toBe('  5');
  });

  it('pads with a custom fill character', () => {
    expect(padLeft('5', 3, '0')).toBe('005');
  });

  it('leaves strings that are already long enough unchanged', () => {
    expect(padLeft('hello', 3)).toBe('hello');
  });
});
