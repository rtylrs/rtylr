import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import { i18next, isRTL } from '@voxire/i18n';

const cookies = vi.hoisted(() => ({
  store: {} as Record<string, string | undefined>,
  setCalls: [] as Array<{ name: string; value: string; options?: Record<string, unknown> }>,
}));

vi.mock('@voxire/auth', () => ({
  getCookie: (name: string) => cookies.store[name] ?? null,
  setCookie: (name: string, value: string, options?: Record<string, unknown>) => {
    cookies.store[name] = value;
    cookies.setCalls.push({ name, value, options });
  },
}));

import { initTestI18n } from '../testing/i18n';
import LanguageSwitch from '../components/LanguageSwitch';
import {
  RENNILI_LANGUAGE_COOKIE,
  applyRenniliLanguage,
  resolveRenniliLanguage,
  selectRenniliLanguage,
} from './renniliLanguage';

beforeAll(async () => {
  await initTestI18n('en');
  i18next.on('languageChanged', (lang: string) => {
    document.documentElement.setAttribute('dir', isRTL(lang) ? 'rtl' : 'ltr');
    document.documentElement.setAttribute('lang', lang);
  });
});

beforeEach(async () => {
  cookies.store = {};
  cookies.setCalls = [];
  await i18next.changeLanguage('en');
});

describe('Rennili language scoping', () => {
  it.each(['/', '/how-it-works', '/features', '/faq', '/rtylr', '/try', '/acme', '/t/tok-1'])(
    'defaults public path %s to Arabic even when Portal has language=en',
    (pathname) => {
      cookies.store.language = 'en';
      expect(resolveRenniliLanguage(pathname)).toBe('ar');
    }
  );

  it('uses Rennili host-only cookie on public paths once the visitor picks a language', () => {
    cookies.store.language = 'en';
    cookies.store[RENNILI_LANGUAGE_COOKIE] = 'fr';
    expect(resolveRenniliLanguage('/')).toBe('fr');
    expect(resolveRenniliLanguage('/acme')).toBe('fr');
  });

  it.each(['/board', '/board/', '/display', '/poster', '/setup', '/insights'])(
    'defaults staff path %s to English even when Rennili public cookie is Arabic',
    (pathname) => {
      cookies.store[RENNILI_LANGUAGE_COOKIE] = 'ar';
      expect(resolveRenniliLanguage(pathname)).toBe('en');
    }
  );

  it('follows the Portal language cookie on staff paths', () => {
    cookies.store.language = 'fr';
    cookies.store[RENNILI_LANGUAGE_COOKIE] = 'ar';
    expect(resolveRenniliLanguage('/board')).toBe('fr');
  });

  it('switches i18next and document direction when moving between public and staff routes', async () => {
    await applyRenniliLanguage('/');
    expect(i18next.language).toBe('ar');
    expect(document.documentElement).toHaveAttribute('lang', 'ar');
    expect(document.documentElement).toHaveAttribute('dir', 'rtl');

    await applyRenniliLanguage('/board');
    expect(i18next.language).toBe('en');
    expect(document.documentElement).toHaveAttribute('lang', 'en');
    expect(document.documentElement).toHaveAttribute('dir', 'ltr');
  });

  it('writes a host-only cookie from selectRenniliLanguage without touching the shared language cookie', () => {
    selectRenniliLanguage('ar');

    expect(cookies.setCalls).toEqual([
      {
        name: RENNILI_LANGUAGE_COOKIE,
        value: 'ar',
        options: { days: 365, sameSite: 'lax' },
      },
    ]);
    expect(cookies.store.language).toBeUndefined();
    expect(cookies.setCalls[0].options).not.toHaveProperty('domain');
  });

  it('keeps LanguageSwitch changes isolated to Rennili', () => {
    cookies.store.language = 'en';
    render(<LanguageSwitch />);

    fireEvent.click(screen.getByRole('button', { name: 'العربية' }));

    expect(cookies.store[RENNILI_LANGUAGE_COOKIE]).toBe('ar');
    expect(cookies.store.language).toBe('en');
  });
});
