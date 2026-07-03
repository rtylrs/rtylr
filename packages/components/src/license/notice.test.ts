import { describe, it, expect } from 'vitest';
import { resolveLicenseNotice } from './notice';
import type { LicenseState } from './types';

const base = (overrides: Partial<LicenseState>): LicenseState => ({
  mode: 'self_hosted',
  valid: false,
  status: 'missing',
  ...overrides,
});

const NOW = Date.parse('2026-06-01T00:00:00Z');

describe('resolveLicenseNotice', () => {
  it('returns nothing for hosted deployments', () => {
    expect(
      resolveLicenseNotice(base({ mode: 'hosted', status: 'hosted', valid: true }), NOW)
    ).toBeNull();
  });

  it('returns nothing for a plainly valid, non-expiring license', () => {
    expect(
      resolveLicenseNotice(
        base({ status: 'valid', valid: true, expiresAt: '2027-01-01T00:00:00Z' }),
        NOW
      )
    ).toBeNull();
  });

  it('returns nothing when there is no state', () => {
    expect(resolveLicenseNotice(undefined, NOW)).toBeNull();
  });

  it('flags expired as an error', () => {
    expect(resolveLicenseNotice(base({ status: 'expired' }), NOW)).toEqual({
      kind: 'expired',
      severity: 'error',
    });
  });

  it('flags invalid and missing as errors', () => {
    expect(resolveLicenseNotice(base({ status: 'invalid' }), NOW)?.severity).toBe('error');
    expect(resolveLicenseNotice(base({ status: 'missing' }), NOW)?.severity).toBe('error');
  });

  it('flags unreachable as a warning', () => {
    expect(resolveLicenseNotice(base({ status: 'unreachable' }), NOW)).toEqual({
      kind: 'unreachable',
      severity: 'warning',
    });
  });

  it('flags a license expiring within 14 days as a warning', () => {
    const expiresAt = '2026-06-10T00:00:00Z'; // 9 days out
    expect(resolveLicenseNotice(base({ status: 'valid', valid: true, expiresAt }), NOW)).toEqual({
      kind: 'expiring',
      severity: 'warning',
      expiresAt,
    });
  });

  it('does not flag a license expiring beyond 14 days', () => {
    expect(
      resolveLicenseNotice(
        base({ status: 'valid', valid: true, expiresAt: '2026-07-01T00:00:00Z' }),
        NOW
      )
    ).toBeNull();
  });

  it('does not flag an already-expired date reported as valid', () => {
    expect(
      resolveLicenseNotice(
        base({ status: 'valid', valid: true, expiresAt: '2026-05-01T00:00:00Z' }),
        NOW
      )
    ).toBeNull();
  });
});
