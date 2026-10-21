import { describe, expect, it } from 'vitest';

import {
  isLegacyVerificationCode,
  isVerificationToken,
  normalizeVerificationToken,
  VERIFICATION_TOKEN_MAX_LENGTH,
} from './verificationToken';

/**
 * These cases are the server's, not the form's: they are taken from
 * `parseEmailVerificationToken` and its table test in
 * services/onion/internal/modules/user/internal/infrastructure/email_verification_test.go.
 * A shape this accepts that the server refuses is a wasted round trip and, on a
 * wrong guess, an attempt charged against the user's budget; a shape this
 * refuses that the server accepts is a user who cannot activate their account.
 */
describe('email verification token', () => {
  describe('normalizeVerificationToken', () => {
    it('keeps the id half, which is what the server needs', () => {
      expect(normalizeVerificationToken('1042.0123456789')).toBe('1042.0123456789');
    });

    it('drops the whitespace a rendered email wraps a copied token in', () => {
      // U+00A0 is not ASCII space, and it is what mail clients actually insert.
      expect(normalizeVerificationToken('  1042.0123456789\n')).toBe('1042.0123456789');
    });

    it('reads the token out of a pasted verification link', () => {
      expect(
        normalizeVerificationToken('https://auth.rtylr.com/verify?token=1042.0123456789')
      ).toBe('1042.0123456789');
      expect(normalizeVerificationToken('http://localhost:5176/verify?token=1042.0123456789')).toBe(
        '1042.0123456789'
      );
    });

    it('finds the parameter in a link a client hard-wrapped across lines', () => {
      // The digit strips cannot save this one: a break inside "token=" hides
      // the parameter from the lookup, so the invisible characters have to go
      // before the link is read.
      expect(normalizeVerificationToken('https://a.test/verify?to\nken=1042.0123 456789')).toBe(
        '1042.0123456789'
      );
    });

    it('stops at the query separator after a pasted link', () => {
      expect(
        normalizeVerificationToken('https://a.test/verify?token=1042.0123456789&email=a@b.c')
      ).toBe('1042.0123456789');
      expect(normalizeVerificationToken('https://a.test/verify?token=1042.0123456789#top')).toBe(
        '1042.0123456789'
      );
    });

    it('does not join digits across a second separator into a code nobody issued', () => {
      expect(normalizeVerificationToken('1042.01234.56789')).toBe('1042.01234');
    });

    it('leaves a bare code without an id rather than inventing one', () => {
      // The email shows the bare ten-digit code as secondary copy, but no
      // client can turn it into a token: there is no id to pair it with.
      expect(normalizeVerificationToken('0123456789')).toBe('0123456789');
    });

    it('keeps a legacy six-digit code intact for the server to refuse', () => {
      expect(normalizeVerificationToken('123456')).toBe('123456');
    });

    it('does not fabricate a leading zero on the id', () => {
      // ParseInt accepts "01042" as 1042; a zero-padded id must survive too.
      expect(normalizeVerificationToken('01042.0123456789')).toBe('01042.0123456789');
    });
  });

  describe('isVerificationToken', () => {
    it('accepts what the server issues', () => {
      expect(isVerificationToken('1042.0123456789')).toBe(true);
      expect(isVerificationToken('1.0000000001')).toBe(true);
      expect(isVerificationToken('999999999.0000000042')).toBe(true);
    });

    it('accepts a leading-zero id, which ParseInt reads as the same number', () => {
      expect(isVerificationToken('01042.0123456789')).toBe(true);
    });

    // Each of these is a "wantParsed: false" row in the server's own table.
    it('refuses exactly the shapes the server cannot parse', () => {
      for (const token of [
        '1042.01234', // too short
        '1042.01234567890', // too long
        '1042.01234567x9', // non-numeric code
        '0.0123456789', // zero id
        '-7.0123456789', // negative id
        'abc.0123456789', // non-numeric id
        '1042.', // no code
        '.0123456789', // no id
        '1042', // no separator
        '0123456789', // no separator
        '',
      ]) {
        expect(isVerificationToken(token), token).toBe(false);
      }
    });

    it('refuses an id past int64, which ParseInt would overflow on', () => {
      expect(isVerificationToken('9223372036854775807.0123456789')).toBe(true);
      expect(isVerificationToken('9223372036854775808.0123456789')).toBe(false);
    });
  });

  describe('isLegacyVerificationCode', () => {
    it('recognises the pre-binding six-digit format', () => {
      expect(isLegacyVerificationCode('123456')).toBe(true);
    });

    it('does not claim a real token is legacy', () => {
      expect(isLegacyVerificationCode('1042.0123456789')).toBe(false);
      expect(isLegacyVerificationCode('12345')).toBe(false);
      expect(isLegacyVerificationCode('1234567')).toBe(false);
    });
  });

  it('caps the field wide enough for the longest token the server issues', () => {
    // 19 digits of int64, the separator, ten digits. A 6-digit cap truncated
    // every real token into something the server rejects as malformed.
    expect(VERIFICATION_TOKEN_MAX_LENGTH).toBe(30);
    expect(isVerificationToken('9223372036854775807.0123456789')).toBe(true);
  });
});
