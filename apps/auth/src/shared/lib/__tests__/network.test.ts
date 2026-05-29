import { describe, it, expect } from 'vitest';
import {
  isSuccess,
  isClientError,
  isServerError,
  isRedirect,
  statusLabel,
  isOnline,
} from '../network';

describe('isSuccess', () => {
  it('returns true for 200', () => expect(isSuccess(200)).toBe(true));
  it('returns true for 204', () => expect(isSuccess(204)).toBe(true));
  it('returns false for 400', () => expect(isSuccess(400)).toBe(false));
});

describe('isClientError', () => {
  it('returns true for 404', () => expect(isClientError(404)).toBe(true));
  it('returns true for 401', () => expect(isClientError(401)).toBe(true));
  it('returns false for 500', () => expect(isClientError(500)).toBe(false));
});

describe('isServerError', () => {
  it('returns true for 500', () => expect(isServerError(500)).toBe(true));
  it('returns true for 503', () => expect(isServerError(503)).toBe(true));
  it('returns false for 200', () => expect(isServerError(200)).toBe(false));
});

describe('isRedirect', () => {
  it('returns true for 301', () => expect(isRedirect(301)).toBe(true));
  it('returns true for 302', () => expect(isRedirect(302)).toBe(true));
  it('returns false for 200', () => expect(isRedirect(200)).toBe(false));
});

describe('statusLabel', () => {
  it('returns OK for 200', () => expect(statusLabel(200)).toBe('OK'));
  it('returns Not Found for 404', () => expect(statusLabel(404)).toBe('Not Found'));
  it('returns a fallback for unknown codes', () => expect(statusLabel(418)).toBe('HTTP 418'));
});

describe('isOnline', () => {
  it('returns a boolean', () => {
    expect(typeof isOnline()).toBe('boolean');
  });
});
