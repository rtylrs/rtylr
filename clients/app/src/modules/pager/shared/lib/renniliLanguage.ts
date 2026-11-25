import { getCookie, setCookie } from '@voxire/auth';
import { i18next, isRTL } from '@voxire/i18n';
import { isRenniliStaffPath } from '../../pagerPaths';

/**
 * Host-only cookie for public Rennili pages (`rennili.rtylr.com`), kept separate
 * from the shared `.rtylr.com` `language` cookie used by Portal/Auth/POS.
 */
export const RENNILI_LANGUAGE_COOKIE = 'rennili_language';

export function resolveRenniliLanguage(pathname: string): string {
  if (isRenniliStaffPath(pathname)) {
    return getCookie('language') || 'en';
  }
  return getCookie(RENNILI_LANGUAGE_COOKIE) || 'ar';
}

export function applyRenniliLanguage(pathname: string): Promise<void> {
  const target = resolveRenniliLanguage(pathname);
  if (i18next.language === target) {
    if (typeof document !== 'undefined') {
      document.documentElement.setAttribute('dir', isRTL(target) ? 'rtl' : 'ltr');
      document.documentElement.setAttribute('lang', target);
    }
    return Promise.resolve();
  }
  return i18next.changeLanguage(target).then(() => undefined);
}

export function selectRenniliLanguage(language: string = 'ar'): void {
  void i18next.changeLanguage(language);
  setCookie(RENNILI_LANGUAGE_COOKIE, language, {
    days: 365,
    sameSite: 'lax',
  });
}
