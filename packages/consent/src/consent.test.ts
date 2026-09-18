import { afterEach, beforeEach, describe, expect, test, vi } from 'vitest';
import { clearConsent, readConsent, writeConsent } from './consent';

/**
 * Tests for the strict-opt-in consent cookie. Verifies the round-trip
 * (write → read), the undecided fallback, and the change-event that
 * downstream packages (`@voxire/observability`) listen on.
 */

function clearConsentCookie(): void {
  document.cookie = 'rtylr-consent=; expires=Thu, 01 Jan 1970 00:00:00 GMT; path=/';
}

describe('consent', () => {
  beforeEach(() => clearConsentCookie());
  afterEach(() => clearConsentCookie());

  test('readConsent returns "undecided" when no cookie is set', () => {
    expect(readConsent()).toBe('undecided');
  });

  test('writeConsent → readConsent round-trips the granted state', () => {
    writeConsent('granted');
    expect(readConsent()).toBe('granted');
  });

  test('writeConsent → readConsent round-trips the denied state', () => {
    writeConsent('denied');
    expect(readConsent()).toBe('denied');
  });

  test('writeConsent fires the rtylr-consent-change window event with raw string detail', () => {
    const listener = vi.fn();
    const wrapped: EventListener = (e) => listener((e as CustomEvent).detail);
    window.addEventListener('rtylr-consent-change', wrapped);
    try {
      writeConsent('granted');
      expect(listener).toHaveBeenCalledWith('granted');
    } finally {
      window.removeEventListener('rtylr-consent-change', wrapped);
    }
  });

  test('clearConsent → readConsent returns "undecided"', () => {
    writeConsent('granted');
    expect(readConsent()).toBe('granted');
    clearConsent();
    expect(readConsent()).toBe('undecided');
  });

  test('clearConsent fires the rtylr-consent-change event with "undecided"', () => {
    writeConsent('granted');
    const listener = vi.fn();
    const wrapped: EventListener = (e) => listener((e as CustomEvent).detail);
    window.addEventListener('rtylr-consent-change', wrapped);
    try {
      clearConsent();
      expect(listener).toHaveBeenCalledWith('undecided');
    } finally {
      window.removeEventListener('rtylr-consent-change', wrapped);
    }
  });

  test('a malformed cookie value falls back to "undecided"', () => {
    document.cookie = `rtylr-consent=${encodeURIComponent('garbage')}; path=/`;
    expect(readConsent()).toBe('undecided');
  });
});
