import { describe, expect, it } from 'vitest';
import { VOICE_LOCALES, DEFAULT_VOICE_LOCALE } from './voiceLocales';

describe('VOICE_LOCALES', () => {
  it('contains the default locale', () => {
    expect(VOICE_LOCALES.some((l) => l.code === DEFAULT_VOICE_LOCALE)).toBe(true);
  });

  it('uses unique codes', () => {
    const codes = VOICE_LOCALES.map((l) => l.code);
    expect(new Set(codes).size).toBe(codes.length);
  });
});
