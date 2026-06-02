import { describe, expect, it } from 'vitest';

import {
  capitalize,
  collapseWhitespace,
  deslugify,
  interpolate,
  readingTimeMinutes,
  slugify,
  stripHtml,
  toTitleCase,
  truncate,
  wordCount,
} from './stringUtils';

describe('truncate', () => {
  it('returns the original string when it is within maxLength', () => {
    expect(truncate('hello', 10)).toBe('hello');
  });

  it('truncates and appends an ellipsis', () => {
    expect(truncate('hello world', 8)).toBe('hello...');
  });

  it('supports a custom ellipsis', () => {
    expect(truncate('hello world', 7, '~')).toBe('hello ~');
  });

  it('returns an empty string when maxLength is 0', () => {
    expect(truncate('hi', 3)).toBe('hi');
  });
});

describe('slugify', () => {
  it('lowercases and replaces spaces with hyphens', () => {
    expect(slugify('Hello World')).toBe('hello-world');
  });

  it('strips special characters', () => {
    expect(slugify('Hello, World!')).toBe('hello-world');
  });

  it('collapses multiple spaces', () => {
    expect(slugify('foo   bar')).toBe('foo-bar');
  });

  it('removes leading and trailing hyphens', () => {
    expect(slugify('  hello  ')).toBe('hello');
  });
});

describe('deslugify', () => {
  it('converts a slug to title case', () => {
    expect(deslugify('hello-world')).toBe('Hello World');
  });

  it('handles single word slugs', () => {
    expect(deslugify('blog')).toBe('Blog');
  });
});

describe('toTitleCase', () => {
  it('title-cases each word', () => {
    expect(toTitleCase('the quick brown fox')).toBe('The Quick Brown Fox');
  });

  it('handles all-caps input', () => {
    expect(toTitleCase('HELLO WORLD')).toBe('Hello World');
  });
});

describe('capitalize', () => {
  it('capitalizes the first character', () => {
    expect(capitalize('hello world')).toBe('Hello world');
  });

  it('returns empty string unchanged', () => {
    expect(capitalize('')).toBe('');
  });
});

describe('interpolate', () => {
  it('replaces named placeholders', () => {
    expect(interpolate('Hello, {name}!', { name: 'World' })).toBe('Hello, World!');
  });

  it('leaves unknown placeholders intact', () => {
    expect(interpolate('Hi {name}', {})).toBe('Hi {name}');
  });

  it('supports numeric values', () => {
    expect(interpolate('Count: {n}', { n: 42 })).toBe('Count: 42');
  });
});

describe('collapseWhitespace', () => {
  it('collapses multiple spaces into one', () => {
    expect(collapseWhitespace('hello   world')).toBe('hello world');
  });

  it('removes leading and trailing whitespace', () => {
    expect(collapseWhitespace('  hello  ')).toBe('hello');
  });

  it('collapses newlines and tabs', () => {
    expect(collapseWhitespace('hello\n\tworld')).toBe('hello world');
  });
});

describe('stripHtml', () => {
  it('removes HTML tags', () => {
    expect(stripHtml('<p>Hello <b>world</b></p>')).toBe('Hello world');
  });

  it('returns plain text unchanged', () => {
    expect(stripHtml('plain text')).toBe('plain text');
  });
});

describe('wordCount', () => {
  it('counts words in a sentence', () => {
    expect(wordCount('hello world foo')).toBe(3);
  });

  it('returns 0 for an empty string', () => {
    expect(wordCount('')).toBe(0);
  });

  it('handles extra whitespace', () => {
    expect(wordCount('  hello   world  ')).toBe(2);
  });
});

describe('readingTimeMinutes', () => {
  it('returns 1 for short text', () => {
    const shortText = 'hello'.repeat(50);
    expect(readingTimeMinutes(shortText)).toBeGreaterThanOrEqual(1);
  });

  it('returns a higher value for longer text', () => {
    const longText = 'word '.repeat(600);
    expect(readingTimeMinutes(longText)).toBeGreaterThan(1);
  });

  it('supports a custom words-per-minute rate', () => {
    const text = 'word '.repeat(100);
    expect(readingTimeMinutes(text, 100)).toBe(1);
  });
});
