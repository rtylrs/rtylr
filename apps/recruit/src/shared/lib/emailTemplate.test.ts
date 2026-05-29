import { describe, it, expect } from 'vitest';
import {
  applicationReceivedSubject,
  statusUpdateSubject,
  candidatePreviewText,
  hrNotificationSubject,
  emailStatusLabel,
  isNotificationStatus,
} from './emailTemplate';

describe('applicationReceivedSubject', () => {
  it('includes the career title and brand name', () => {
    const result = applicationReceivedSubject('Senior Engineer', 'Acme Corp');
    expect(result).toContain('Senior Engineer');
    expect(result).toContain('Acme Corp');
  });
});

describe('statusUpdateSubject', () => {
  it('includes the career title and brand name', () => {
    const result = statusUpdateSubject('Frontend Developer', 'Acme');
    expect(result).toContain('Frontend Developer');
    expect(result).toContain('Acme');
  });
});

describe('candidatePreviewText', () => {
  it('uses the first name from the candidate name', () => {
    const result = candidatePreviewText('Alice Smith', 'Acme');
    expect(result).toContain('Alice');
    expect(result).not.toContain('Smith');
  });

  it('includes the brand name', () => {
    const result = candidatePreviewText('Bob', 'Globex');
    expect(result).toContain('Globex');
  });

  it('handles a single-word name', () => {
    const result = candidatePreviewText('Bob', 'Acme');
    expect(result).toContain('Bob');
  });
});

describe('hrNotificationSubject', () => {
  it('includes career title and candidate name', () => {
    const result = hrNotificationSubject('Backend Engineer', 'Alice Smith');
    expect(result).toContain('Backend Engineer');
    expect(result).toContain('Alice Smith');
  });
});

describe('emailStatusLabel', () => {
  it('returns "received" for pending', () => {
    expect(emailStatusLabel('pending')).toBe('received');
  });

  it('returns "shortlisted" for shortlisted', () => {
    expect(emailStatusLabel('shortlisted')).toBe('shortlisted');
  });

  it('returns "selected" for hired', () => {
    expect(emailStatusLabel('hired')).toBe('selected');
  });

  it('returns the raw status for unknown values', () => {
    expect(emailStatusLabel('archived')).toBe('archived');
  });
});

describe('isNotificationStatus', () => {
  it('returns true for reviewing', () => {
    expect(isNotificationStatus('reviewing')).toBe(true);
  });

  it('returns true for hired', () => {
    expect(isNotificationStatus('hired')).toBe(true);
  });

  it('returns true for rejected', () => {
    expect(isNotificationStatus('rejected')).toBe(true);
  });

  it('returns false for pending (initial confirmation is separate)', () => {
    expect(isNotificationStatus('pending')).toBe(false);
  });

  it('returns false for unknown statuses', () => {
    expect(isNotificationStatus('archived')).toBe(false);
  });
});
