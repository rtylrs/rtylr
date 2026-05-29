import { describe, it, expect } from 'vitest';
import {
  hasAllConsents,
  hasAnyConsent,
  consentedChannels,
  canContact,
  consentSummary,
  emptyConsent,
} from './consentUtils';
const full = { email: true, sms: true, marketing: true };
const partial = { email: true, sms: false, marketing: false };
const none = { email: false, sms: false, marketing: false };
describe('hasAllConsents', () => {
  it('all', () => expect(hasAllConsents(full)).toBe(true));
  it('partial', () => expect(hasAllConsents(partial)).toBe(false));
});
describe('hasAnyConsent', () => {
  it('some', () => expect(hasAnyConsent(partial)).toBe(true));
  it('none', () => expect(hasAnyConsent(none)).toBe(false));
});
describe('consentedChannels', () => {
  it('partial', () => expect(consentedChannels(partial)).toEqual(['email']));
  it('empty', () => expect(consentedChannels(none)).toEqual([]));
});
describe('canContact', () => {
  it('email ok', () => expect(canContact(partial, 'email')).toBe(true));
  it('sms not ok', () => expect(canContact(partial, 'sms')).toBe(false));
});
describe('consentSummary', () => {
  it('has channels', () => expect(consentSummary(full)).toContain('email'));
  it('no consents', () => expect(consentSummary(none)).toBe('No consents'));
});
describe('emptyConsent', () => {
  it('all false', () => {
    const c = emptyConsent();
    expect(c.email).toBe(false);
  });
});
