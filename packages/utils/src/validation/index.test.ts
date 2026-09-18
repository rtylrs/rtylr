import { describe, it, expect } from 'vitest';
import { validation } from './index';

describe('validation', () => {
  // ── Email ──────────────────────────────────────────────────────

  describe('email', () => {
    it('validates correct email', () => {
      expect(validation.email('test@example.com')).toBe(true);
    });

    it('validates email with subdomain', () => {
      expect(validation.email('test@mail.example.com')).toBe(true);
    });

    it('rejects email without @', () => {
      expect(validation.email('testexample.com')).toBe(false);
    });

    it('rejects email without domain', () => {
      expect(validation.email('test@')).toBe(false);
    });

    it('rejects email with spaces', () => {
      expect(validation.email('test @example.com')).toBe(false);
    });

    it('rejects empty string', () => {
      expect(validation.email('')).toBe(false);
    });
  });

  // ── Phone ──────────────────────────────────────────────────────

  describe('phone', () => {
    it('validates phone number', () => {
      expect(validation.phone('1234567890')).toBe(true);
    });

    it('validates phone with country code', () => {
      expect(validation.phone('+1234567890')).toBe(true);
    });

    it('validates phone with spaces', () => {
      expect(validation.phone('+1 234 567 890')).toBe(true);
    });

    it('rejects phone starting with 0', () => {
      expect(validation.phone('0123456789')).toBe(false);
    });

    it('rejects empty string', () => {
      expect(validation.phone('')).toBe(false);
    });

    it('rejects letters', () => {
      expect(validation.phone('abc')).toBe(false);
    });
  });

  // ── Password ──────────────────────────────────────────────────

  describe('password', () => {
    it('validates strong password', () => {
      const result = validation.password('StrongP@ss1');
      expect(result.isValid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    it('rejects short password', () => {
      const result = validation.password('Sh@1');
      expect(result.isValid).toBe(false);
      expect(result.errors).toContain('Password must be at least 8 characters long');
    });

    it('requires uppercase letter', () => {
      const result = validation.password('lowercase@1');
      expect(result.isValid).toBe(false);
      expect(result.errors).toContain('Password must contain at least one uppercase letter');
    });

    it('requires lowercase letter', () => {
      const result = validation.password('UPPERCASE@1');
      expect(result.isValid).toBe(false);
      expect(result.errors).toContain('Password must contain at least one lowercase letter');
    });

    it('requires number', () => {
      const result = validation.password('NoNumber@s');
      expect(result.isValid).toBe(false);
      expect(result.errors).toContain('Password must contain at least one number');
    });

    it('requires special character', () => {
      const result = validation.password('NoSpecial1a');
      expect(result.isValid).toBe(false);
      expect(result.errors).toContain('Password must contain at least one special character');
    });

    it('returns multiple errors for weak password', () => {
      const result = validation.password('abc');
      expect(result.isValid).toBe(false);
      expect(result.errors.length).toBeGreaterThan(1);
    });

    it('returns all 5 errors for empty-like password', () => {
      const result = validation.password('');
      expect(result.errors).toHaveLength(5);
    });
  });

  // ── Required ──────────────────────────────────────────────────

  describe('required', () => {
    it('returns true for non-empty string', () => {
      expect(validation.required('hello')).toBe(true);
    });

    it('returns false for empty string', () => {
      expect(validation.required('')).toBe(false);
    });

    it('returns false for whitespace-only string', () => {
      expect(validation.required('   ')).toBe(false);
    });

    it('returns false for null', () => {
      expect(validation.required(null)).toBe(false);
    });

    it('returns false for undefined', () => {
      expect(validation.required(undefined)).toBe(false);
    });

    it('returns true for number 0', () => {
      expect(validation.required(0)).toBe(true);
    });

    it('returns true for false boolean', () => {
      expect(validation.required(false)).toBe(true);
    });

    it('returns true for object', () => {
      expect(validation.required({})).toBe(true);
    });
  });

  // ── minLength / maxLength ──────────────────────────────────────

  describe('minLength', () => {
    it('returns true when string meets min', () => {
      expect(validation.minLength('hello', 3)).toBe(true);
    });

    it('returns true at exact min', () => {
      expect(validation.minLength('abc', 3)).toBe(true);
    });

    it('returns false when too short', () => {
      expect(validation.minLength('ab', 3)).toBe(false);
    });

    it('handles empty string', () => {
      expect(validation.minLength('', 1)).toBe(false);
    });
  });

  describe('maxLength', () => {
    it('returns true when string is under max', () => {
      expect(validation.maxLength('hi', 5)).toBe(true);
    });

    it('returns true at exact max', () => {
      expect(validation.maxLength('hello', 5)).toBe(true);
    });

    it('returns false when too long', () => {
      expect(validation.maxLength('hello world', 5)).toBe(false);
    });
  });

  // ── URL ──────────────────────────────────────────────────────

  describe('url', () => {
    it('validates https URL', () => {
      expect(validation.url('https://example.com')).toBe(true);
    });

    it('validates http URL', () => {
      expect(validation.url('http://example.com')).toBe(true);
    });

    it('rejects invalid URL', () => {
      expect(validation.url('not-a-url')).toBe(false);
    });

    it('rejects empty string', () => {
      expect(validation.url('')).toBe(false);
    });
  });

  // ── Numeric / Alphanumeric ──────────────────────────────────────

  describe('numeric', () => {
    it('returns true for digit string', () => {
      expect(validation.numeric('12345')).toBe(true);
    });

    it('returns false for letters', () => {
      expect(validation.numeric('abc')).toBe(false);
    });

    it('returns false for mixed', () => {
      expect(validation.numeric('123abc')).toBe(false);
    });

    it('returns false for empty string', () => {
      expect(validation.numeric('')).toBe(false);
    });

    it('returns false for decimal', () => {
      expect(validation.numeric('12.5')).toBe(false);
    });
  });

  describe('alphanumeric', () => {
    it('returns true for alphanumeric', () => {
      expect(validation.alphanumeric('abc123')).toBe(true);
    });

    it('returns true for letters only', () => {
      expect(validation.alphanumeric('abcdef')).toBe(true);
    });

    it('returns true for numbers only', () => {
      expect(validation.alphanumeric('123456')).toBe(true);
    });

    it('returns false for special chars', () => {
      expect(validation.alphanumeric('abc@123')).toBe(false);
    });

    it('returns false for spaces', () => {
      expect(validation.alphanumeric('hello world')).toBe(false);
    });

    it('returns false for empty string', () => {
      expect(validation.alphanumeric('')).toBe(false);
    });
  });
});
