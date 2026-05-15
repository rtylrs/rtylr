import { describe, expect, it } from 'vitest';
import { normalizeLang, prettyLang } from './codeLang';

describe('normalizeLang', () => {
  it('rewrites js → javascript', () => {
    expect(normalizeLang('js')).toBe('javascript');
  });
  it('returns undefined for empty input', () => {
    expect(normalizeLang(undefined)).toBeUndefined();
  });
  it('lowercases unknown values', () => {
    expect(normalizeLang('Erlang')).toBe('erlang');
  });
});

describe('prettyLang', () => {
  it('formats known names', () => {
    expect(prettyLang('ts')).toBe('TypeScript');
  });
  it('falls back to the lowercase normalized value', () => {
    expect(prettyLang('erlang')).toBe('erlang');
  });
  it('returns "code" for empty input', () => {
    expect(prettyLang(undefined)).toBe('code');
  });
});
