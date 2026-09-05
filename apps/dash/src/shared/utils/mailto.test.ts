import { describe, it, expect } from 'vitest';
import { mailtoFor } from './mailto';

describe('mailtoFor', () => {
  it('encodes valid addresses', () => {
    expect(mailtoFor('jane@rtylr.com')).toBe('mailto:jane%40rtylr.com');
    expect(mailtoFor('  jane@rtylr.com  ')).toBe('mailto:jane%40rtylr.com');
  });

  it('neutralizes query and header injection', () => {
    expect(mailtoFor('a@b.com?subject=x&body=y')).toBe('mailto:a%40b.com%3Fsubject%3Dx%26body%3Dy');
    expect(mailtoFor('a@b.com\nBcc: evil@x.com')).toBeNull();
  });

  it('rejects missing and malformed values', () => {
    expect(mailtoFor(null)).toBeNull();
    expect(mailtoFor(undefined)).toBeNull();
    expect(mailtoFor('')).toBeNull();
    expect(mailtoFor('not-an-email')).toBeNull();
    expect(mailtoFor('a@b')).toBeNull();
    expect(mailtoFor(42)).toBeNull();
  });
});
