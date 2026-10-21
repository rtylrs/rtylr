import { describe, expect, it } from 'vitest';

import en from '../locales/en.json';
import type { FaqSection } from './faqs';

/**
 * The FAQ copy lives in the locale files, so these assert the editorial rules
 * documented at the top of lib/faqs.ts against the English source of truth.
 * The old version of this file read the seven module-scope arrays the copy
 * moved out of; it tested a shape that no longer exists (rtylr#1361).
 */

const SECTIONS: FaqSection[] = [
  'home',
  'pricing',
  'workspace',
  'demo',
];

interface FaqEntry {
  question: string;
  answer: string;
}

function items(section: FaqSection): FaqEntry[] {
  return (en.faq as Record<string, { items: FaqEntry[] }>)[section].items;
}

describe('FAQ structure', () => {
  it.each(SECTIONS)('%s has a non-empty item list', (section) => {
    expect(items(section).length).toBeGreaterThan(0);
  });

  it.each(SECTIONS)('%s questions are questions and answers are answers', (section) => {
    for (const entry of items(section)) {
      expect(entry.question.trim().endsWith('?'), entry.question).toBe(true);
      expect(entry.answer.trim().length).toBeGreaterThan(0);
    }
  });

  it.each(SECTIONS)('%s carries no em dash or hyphen standing in for one', (section) => {
    for (const entry of items(section)) {
      expect(entry.answer).not.toContain('—');
      expect(entry.answer).not.toContain(' - ');
      expect(entry.question).not.toContain('—');
      expect(entry.question).not.toContain(' - ');
    }
  });

  it.each(SECTIONS)('%s publishes no price', (section) => {
    for (const entry of items(section)) {
      expect(entry.answer).not.toMatch(/[$€£]\s?\d/);
      expect(entry.answer.toLowerCase()).not.toMatch(/\b\d+\s?(usd|eur|gbp)\b/);
    }
  });

  it('answers are long enough to survive snippet extraction', () => {
    // The rule in lib/faqs.ts is roughly 40 to 80 words. The pricing section
    // is deliberately shorter (the plan data is the authority on pricing), so
    // the floor here is the floor that holds for every section.
    for (const section of SECTIONS) {
      for (const entry of items(section)) {
        const words = entry.answer.trim().split(/\s+/).length;
        expect(words, `${section}: ${entry.question}`).toBeGreaterThanOrEqual(15);
      }
    }
  });
});
