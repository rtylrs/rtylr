import { describe, it, expect } from 'vitest';
import {
  validateRequired,
  validateEmail,
  validatePhone,
  validateUrl,
  validateMaxLength,
  validateFileSize,
  validateFileType,
} from './validation';

describe('validateRequired', () => {
  it('returns null for a non-empty string', () => {
    expect(validateRequired('hello')).toBeNull();
  });

  it('returns an error for an empty string', () => {
    expect(validateRequired('')).not.toBeNull();
  });

  it('returns an error for a whitespace-only string', () => {
    expect(validateRequired('   ')).not.toBeNull();
  });

  it('includes the label in the error message', () => {
    const err = validateRequired('', 'Name');
    expect(err).toContain('Name');
  });
});

describe('validateEmail', () => {
  it('returns null for a valid email', () => {
    expect(validateEmail('user@example.com')).toBeNull();
  });

  it('returns an error for missing @', () => {
    expect(validateEmail('notanemail')).not.toBeNull();
  });

  it('returns an error for @ at the start', () => {
    expect(validateEmail('@example.com')).not.toBeNull();
  });

  it('returns an error for missing TLD', () => {
    expect(validateEmail('user@example')).not.toBeNull();
  });

  it('returns an error for empty string', () => {
    expect(validateEmail('')).not.toBeNull();
  });
});

describe('validatePhone', () => {
  it('returns null for an empty string (optional field)', () => {
    expect(validatePhone('')).toBeNull();
  });

  it('returns null for a valid phone number', () => {
    expect(validatePhone('+1 (555) 123-4567')).toBeNull();
  });

  it('returns an error for a string with too few digits', () => {
    expect(validatePhone('123')).not.toBeNull();
  });

  it('returns null for a 7-digit number', () => {
    expect(validatePhone('5551234')).toBeNull();
  });
});

describe('validateUrl', () => {
  it('returns null for an empty string (optional field)', () => {
    expect(validateUrl('')).toBeNull();
  });

  it('returns null for a valid https URL', () => {
    expect(validateUrl('https://example.com/path')).toBeNull();
  });

  it('returns null for a valid http URL', () => {
    expect(validateUrl('http://example.com')).toBeNull();
  });

  it('returns an error for a non-URL string', () => {
    expect(validateUrl('not a url')).not.toBeNull();
  });

  it('returns an error for a ftp URL', () => {
    expect(validateUrl('ftp://example.com')).not.toBeNull();
  });
});

describe('validateMaxLength', () => {
  it('returns null when within limit', () => {
    expect(validateMaxLength('hello', 10)).toBeNull();
  });

  it('returns null when exactly at the limit', () => {
    expect(validateMaxLength('hello', 5)).toBeNull();
  });

  it('returns an error when over the limit', () => {
    expect(validateMaxLength('hello world', 5)).not.toBeNull();
  });

  it('includes max in the error message', () => {
    const err = validateMaxLength('toolong', 5);
    expect(err).toContain('5');
  });
});

describe('validateFileSize', () => {
  function makeFile(size: number): File {
    return new File([new ArrayBuffer(size)], 'test.pdf', { type: 'application/pdf' });
  }

  it('returns null when file is within limit', () => {
    expect(validateFileSize(makeFile(1024), 5 * 1024 * 1024)).toBeNull();
  });

  it('returns an error when file exceeds limit', () => {
    expect(validateFileSize(makeFile(10 * 1024 * 1024), 5 * 1024 * 1024)).not.toBeNull();
  });

  it('includes MB limit in the error message', () => {
    const err = validateFileSize(makeFile(10 * 1024 * 1024), 5 * 1024 * 1024);
    expect(err).toContain('5 MB');
  });
});

describe('validateFileType', () => {
  function makeFile(type: string): File {
    return new File([''], 'test', { type });
  }

  it('returns null when file type is accepted', () => {
    expect(validateFileType(makeFile('application/pdf'), ['application/pdf'])).toBeNull();
  });

  it('returns an error for a rejected type', () => {
    expect(validateFileType(makeFile('image/gif'), ['application/pdf'])).not.toBeNull();
  });
});
