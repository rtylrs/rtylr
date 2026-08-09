/**
 * Auth app constants.
 * Centralises magic strings and numbers used across features.
 */

/** Cookie name for the session token. */
export const TOKEN_COOKIE_NAME = 'token';

/** Number of days before the session cookie expires. */
export const TOKEN_COOKIE_DAYS = 30;

/** Minimum password length enforced by the UI. */
export const MIN_PASSWORD_LENGTH = 8;

/** Minimum password strength score (0-4) required to submit the register form. */
export const MIN_PASSWORD_SCORE = 2;

/** Local storage key used to preserve the OAuth redirect destination. */
export const GOOGLE_AUTH_REDIRECT_KEY = 'google_auth_redirect';

/** Stale time (in ms) for the subscription plans query. */
export const PLANS_STALE_TIME_MS = 5 * 60 * 1000;

/** App IDs that have a matching solo plan. */
export const SOLO_APP_IDS = ['flow', 'finance', 'crm', 'hr', 'erp', 'pos'] as const;

export type SoloAppId = (typeof SOLO_APP_IDS)[number];

/** Maps an app redirect param to its solo plan slug. */
export const APP_PLAN_MAP: Record<SoloAppId, string> = {
  flow: 'flow_solo',
  finance: 'finance_solo',
  crm: 'crm_solo',
  hr: 'hr_solo',
  erp: 'erp_solo',
  pos: 'pos_solo',
};

/** Generic error message shown on login failure (avoids user enumeration). */
export const GENERIC_LOGIN_ERROR = 'auth.login.genericError';
