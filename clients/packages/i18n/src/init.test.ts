import { describe, it, expect, vi, beforeEach } from 'vitest';
import { initI18n } from './init';
import { isRTL, getLanguage } from './languages';

describe('i18n package', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('correctly identifies RTL languages', () => {
    expect(isRTL('ar')).toBe(true);
    expect(isRTL('en')).toBe(false);
    expect(isRTL('fr')).toBe(false);
  });

  it('resolves direction for a regional tag, which an exact match missed', () => {
    expect(isRTL('ar-EG')).toBe(true);
    expect(isRTL('ar-SA')).toBe(true);
    expect(isRTL('he-IL')).toBe(true);
    expect(isRTL('ur-PK')).toBe(true);
    expect(isRTL('en-GB')).toBe(false);
    expect(isRTL('fr-CA')).toBe(false);
    expect(isRTL('zh-CN')).toBe(false);
  });

  it('returns native language names', () => {
    expect(getLanguage('ar')?.nativeName).toBe('العربية');
    expect(getLanguage('en')?.nativeName).toBe('English');
    expect(getLanguage('fr')?.nativeName).toBe('Français');
  });

  it('initializes i18n safely without document crash', () => {
    const instance = initI18n({
      loadResource: () => Promise.resolve({ test: 'value' }),
    });
    expect(instance).not.toBeNull();
  });
});
