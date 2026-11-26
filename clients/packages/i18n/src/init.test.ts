import { describe, it, expect, vi, beforeEach } from 'vitest';
import * as auth from '@voxire/auth';
import { initI18n, selectLanguage } from './init';
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

  it('clears any stale host-only language cookie before writing the shared domain cookie', () => {
    vi.spyOn(auth, 'getDefaultCookieDomain').mockReturnValue('.rtylr.com');
    const removeSpy = vi.spyOn(auth, 'removeCookie').mockImplementation(() => {});
    const setSpy = vi.spyOn(auth, 'setCookie').mockImplementation(() => {});

    selectLanguage('en');

    expect(removeSpy).toHaveBeenCalledWith('language');
    expect(setSpy).toHaveBeenCalledWith('language', 'en', {
      days: 365,
      sameSite: 'lax',
      domain: '.rtylr.com',
    });
    expect(removeSpy.mock.invocationCallOrder[0]).toBeLessThan(
      setSpy.mock.invocationCallOrder[0]
    );
  });

  it('does not remove host-only cookie on localhost when no shared domain applies', () => {
    vi.spyOn(auth, 'getDefaultCookieDomain').mockReturnValue(undefined);
    const removeSpy = vi.spyOn(auth, 'removeCookie').mockImplementation(() => {});
    const setSpy = vi.spyOn(auth, 'setCookie').mockImplementation(() => {});

    selectLanguage('en');

    expect(removeSpy).not.toHaveBeenCalled();
    expect(setSpy).toHaveBeenCalledWith('language', 'en', {
      days: 365,
      sameSite: 'lax',
      domain: undefined,
    });
  });
});
