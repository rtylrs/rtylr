import fs from 'node:fs';
import path from 'node:path';
import { describe, expect, it } from 'vitest';
import { MATCH_WORDS, matchName, matchOf, otherWordLanguage, wordIn } from './matchWords';

const GO_WORDS = path.resolve(
  __dirname,
  '../../../../../../../services/onion/internal/modules/pager/internal/domain/match_word.go'
);

/** The keys in the service's MatchWords, in order. */
function serviceKeys(): string[] {
  const source = fs.readFileSync(GO_WORDS, 'utf8');
  const list = /var MatchWords = \[\]string\{([\s\S]*?)\n\}/.exec(source)?.[1] ?? '';
  return [...list.matchAll(/"([^"]*)"/g)].map((found) => found[1] ?? '');
}

describe('match words', () => {
  // The service picks the key and the colour from its position; the phone and
  // board show the emoji and word for it. A key missing here shows as raw text.
  it('has the same keys in the same order as the pager service', () => {
    const keys = serviceKeys();
    expect(keys.length).toBeGreaterThan(70);
    expect(MATCH_WORDS.map((entry) => entry.key)).toEqual(keys);
  });

  it('gives every word an emoji and a distinct word in each language', () => {
    for (const language of ['en', 'fr', 'ar'] as const) {
      const words = MATCH_WORDS.map((entry) => entry[language].toLowerCase());
      expect(new Set(words).size).toBe(words.length);
    }
    const emoji = MATCH_WORDS.map((entry) => entry.emoji);
    expect(emoji.every(Boolean)).toBe(true);
    expect(new Set(emoji).size).toBe(emoji.length);
  });

  it('shows the word in the page language with its number', () => {
    const panda = matchOf({ word: 'panda', wordNumber: 2 });
    if (!panda) throw new Error('panda is a word');
    expect(wordIn(panda, 'en')).toBe('Panda 2');
    expect(wordIn(panda, 'ar')).toBe('باندا 2');
    expect(matchName(matchOf({ word: 'slay' })!, 'fr')).toBe('💅 Slay');
    expect(otherWordLanguage('ar')).toBe('en');
    expect(otherWordLanguage('fr-FR')).toBe('ar');
  });

  it('still shows a two character code and a key this build does not know', () => {
    expect(wordIn(matchOf({ code: 'A7' })!, 'en')).toBe('A7');
    expect(matchName(matchOf({ word: 'zebra' })!, 'en')).toBe('zebra');
    expect(matchOf({})).toBeNull();
  });
});
