import { describe, it, expect } from 'vitest';
import {
  TOKEN_COOKIE_NAME,
  TOKEN_COOKIE_DAYS,
  MIN_PASSWORD_LENGTH,
  MIN_PASSWORD_SCORE,
  GOOGLE_AUTH_REDIRECT_KEY,
  PLANS_STALE_TIME_MS,
  SOLO_APP_IDS,
  APP_PLAN_MAP,
  GENERIC_LOGIN_ERROR,
} from '../constants';

describe('auth constants', () => {
  it('TOKEN_COOKIE_NAME is "token"', () => {
    expect(TOKEN_COOKIE_NAME).toBe('token');
  });

  it('TOKEN_COOKIE_DAYS is 30', () => {
    expect(TOKEN_COOKIE_DAYS).toBe(30);
  });

  it('MIN_PASSWORD_LENGTH is 8', () => {
    expect(MIN_PASSWORD_LENGTH).toBe(8);
  });

  it('MIN_PASSWORD_SCORE is 2', () => {
    expect(MIN_PASSWORD_SCORE).toBe(2);
  });

  it('GOOGLE_AUTH_REDIRECT_KEY is a non-empty string', () => {
    expect(typeof GOOGLE_AUTH_REDIRECT_KEY).toBe('string');
    expect(GOOGLE_AUTH_REDIRECT_KEY.length).toBeGreaterThan(0);
  });

  it('PLANS_STALE_TIME_MS is 5 minutes in milliseconds', () => {
    expect(PLANS_STALE_TIME_MS).toBe(5 * 60 * 1000);
  });

  it('SOLO_APP_IDS includes expected apps', () => {
    expect(SOLO_APP_IDS).toContain('pos');
    expect(SOLO_APP_IDS).toContain('crm');
    expect(SOLO_APP_IDS).toContain('hr');
  });

  it('APP_PLAN_MAP has an entry for every SOLO_APP_ID', () => {
    for (const appId of SOLO_APP_IDS) {
      expect(APP_PLAN_MAP[appId]).toBeTruthy();
      expect(APP_PLAN_MAP[appId]).toContain('_solo');
    }
  });

  it('GENERIC_LOGIN_ERROR is an i18n key string', () => {
    expect(typeof GENERIC_LOGIN_ERROR).toBe('string');
    expect(GENERIC_LOGIN_ERROR.startsWith('auth.')).toBe(true);
  });
});
