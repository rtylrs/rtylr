import { describe, it, expect } from 'vitest';
import {
  EMAIL_RE,
  STRONG_PASSWORD_RE,
  PHONE_RE,
  UUID_V4_RE,
  URL_RE,
  SLUG_RE,
  HEX_COLOR_RE,
  matches,
} from '../regex';

describe('EMAIL_RE', () => {
  it('matches valid email', () => {
    expect(EMAIL_RE.test('user@example.com')).toBe(true);
  });

  it('rejects invalid email', () => {
    expect(EMAIL_RE.test('notanemail')).toBe(false);
  });
});

describe('STRONG_PASSWORD_RE', () => {
  it('matches a strong password', () => {
    expect(STRONG_PASSWORD_RE.test('Abcdef1!')).toBe(true);
  });

  it('rejects a weak password', () => {
    expect(STRONG_PASSWORD_RE.test('password')).toBe(false);
  });
});

describe('PHONE_RE', () => {
  it('matches a phone number', () => {
    expect(PHONE_RE.test('+1 (555) 123-4567')).toBe(true);
  });

  it('rejects non-phone strings', () => {
    expect(PHONE_RE.test('abc')).toBe(false);
  });
});

describe('UUID_V4_RE', () => {
  it('matches a v4 UUID', () => {
    expect(UUID_V4_RE.test('f47ac10b-58cc-4372-a567-0e02b2c3d479')).toBe(true);
  });

  it('rejects non-UUID strings', () => {
    expect(UUID_V4_RE.test('hello-world')).toBe(false);
  });
});

describe('URL_RE', () => {
  it('matches an http URL', () => {
    expect(URL_RE.test('https://example.com')).toBe(true);
  });

  it('rejects plain text', () => {
    expect(URL_RE.test('not a url')).toBe(false);
  });
});

describe('SLUG_RE', () => {
  it('matches a valid slug', () => {
    expect(SLUG_RE.test('my-page-slug')).toBe(true);
  });

  it('rejects slugs with uppercase', () => {
    expect(SLUG_RE.test('My-Slug')).toBe(false);
  });
});

describe('HEX_COLOR_RE', () => {
  it('matches 6-digit hex', () => {
    expect(HEX_COLOR_RE.test('#ff5733')).toBe(true);
  });

  it('matches 3-digit hex', () => {
    expect(HEX_COLOR_RE.test('#fff')).toBe(true);
  });

  it('rejects invalid hex', () => {
    expect(HEX_COLOR_RE.test('red')).toBe(false);
  });
});

describe('matches', () => {
  it('returns true for a match', () => {
    expect(matches('hello@example.com', EMAIL_RE)).toBe(true);
  });

  it('returns false for no match', () => {
    expect(matches('notanemail', EMAIL_RE)).toBe(false);
  });
});
