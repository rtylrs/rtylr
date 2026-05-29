import { describe, it, expect } from 'vitest';
import { parseToc, filterTocDepth, hasToc, tocTitle } from '../toc';

const SAMPLE_MD = `
# Title

## Introduction

Some text here.

## Getting Started

### Installation

### Configuration

## Advanced Usage
`.trim();

describe('parseToc', () => {
  it('extracts h2 and h3 headings', () => {
    const toc = parseToc(SAMPLE_MD);
    expect(toc.length).toBe(5);
  });

  it('ignores h1 headings', () => {
    const toc = parseToc(SAMPLE_MD);
    expect(toc.every((e) => e.level >= 2)).toBe(true);
  });

  it('generates slugs', () => {
    const toc = parseToc(SAMPLE_MD);
    expect(toc[0].slug).toBe('introduction');
  });

  it('returns empty array for no headings', () => {
    expect(parseToc('plain text')).toEqual([]);
  });
});

describe('filterTocDepth', () => {
  it('filters to h2 only', () => {
    const toc = parseToc(SAMPLE_MD);
    const filtered = filterTocDepth(toc, 2);
    expect(filtered.every((e) => e.level === 2)).toBe(true);
  });

  it('includes h3 when maxDepth is 3', () => {
    const toc = parseToc(SAMPLE_MD);
    const filtered = filterTocDepth(toc, 3);
    expect(filtered.some((e) => e.level === 3)).toBe(true);
  });
});

describe('hasToc', () => {
  it('returns true for non-empty TOC', () => {
    expect(hasToc(parseToc(SAMPLE_MD))).toBe(true);
  });

  it('returns false for empty TOC', () => {
    expect(hasToc([])).toBe(false);
  });
});

describe('tocTitle', () => {
  it('returns first heading title', () => {
    expect(tocTitle(parseToc(SAMPLE_MD))).toBe('Introduction');
  });

  it('returns undefined for empty toc', () => {
    expect(tocTitle([])).toBeUndefined();
  });
});
