import { describe, it, expect } from 'vitest';
import { emailTemplates } from './emailTemplates';

// ─── emailTemplates - pure function unit tests ────────────────────────────────
//
// These templates have zero external dependencies: they take a plain context
// object and return { subject, body }. No mocks are needed.

describe('emailTemplates.birthday', () => {
  it('uses the provided name in subject and body', () => {
    const result = emailTemplates.birthday({ name: 'Alice' });
    expect(result.subject).toContain('Alice');
    expect(result.body).toContain('Alice');
  });

  it('defaults company to "rtylr" in the body', () => {
    const result = emailTemplates.birthday({ name: 'Alice' });
    // The birthday subject only contains the recipient's name - company
    // appears only in the body and closing signature.
    expect(result.body).toContain('rtylr');
  });

  it('uses a custom company name when provided', () => {
    const result = emailTemplates.birthday({ name: 'Bob', company: 'AcmeCorp' });
    expect(result.body).toContain('AcmeCorp');
    expect(result.body).not.toContain('rtylr');
  });

  it('returns both subject and body fields', () => {
    const result = emailTemplates.birthday({ name: 'Alice' });
    expect(result).toHaveProperty('subject');
    expect(result).toHaveProperty('body');
  });

  it('subject says "Happy birthday"', () => {
    const result = emailTemplates.birthday({ name: 'Eve' });
    expect(result.subject).toMatch(/happy birthday/i);
  });
});

describe('emailTemplates.anniversary', () => {
  it('uses the provided name in subject and body', () => {
    const result = emailTemplates.anniversary({ name: 'Charlie', years: 3 });
    expect(result.subject).toContain('Charlie');
    expect(result.body).toContain('Charlie');
  });

  it('reflects the correct year count', () => {
    const result = emailTemplates.anniversary({ name: 'Dana', years: 5 });
    expect(result.subject).toContain('5');
    expect(result.body).toContain('5');
  });

  it('uses singular "year" for exactly 1 year', () => {
    const result = emailTemplates.anniversary({ name: 'Eve', years: 1 });
    expect(result.subject).toContain('1 year');
    expect(result.subject).not.toContain('1 years');
    expect(result.body).toContain('1 year');
    expect(result.body).not.toContain('1 years');
  });

  it('uses plural "years" for more than 1 year', () => {
    const result = emailTemplates.anniversary({ name: 'Frank', years: 2 });
    expect(result.subject).toContain('2 years');
    expect(result.body).toContain('2 years');
  });

  it('defaults years to 1 when not provided', () => {
    const result = emailTemplates.anniversary({ name: 'Grace' });
    expect(result.subject).toContain('1 year');
  });

  it('defaults company to "rtylr"', () => {
    const result = emailTemplates.anniversary({ name: 'Harry', years: 2 });
    expect(result.body).toContain('rtylr');
  });

  it('uses a custom company name when provided', () => {
    const result = emailTemplates.anniversary({ name: 'Iris', years: 4, company: 'VoxireCo' });
    expect(result.body).toContain('VoxireCo');
    expect(result.body).not.toContain('rtylr');
  });
});

describe('emailTemplates.documentExpiring', () => {
  it('uses the provided name in the body', () => {
    const result = emailTemplates.documentExpiring({
      name: 'Jake',
      documentName: 'Passport',
      expiresOn: '2026-06-01',
    });
    expect(result.body).toContain('Jake');
  });

  it('includes the document name in subject and body', () => {
    const result = emailTemplates.documentExpiring({
      name: 'Karen',
      documentName: 'Work Permit',
      expiresOn: '2026-07-15',
    });
    expect(result.subject).toContain('Work Permit');
    expect(result.body).toContain('Work Permit');
  });

  it('includes the expiration date in the body', () => {
    const result = emailTemplates.documentExpiring({
      name: 'Leo',
      documentName: 'Contract',
      expiresOn: '2026-12-31',
    });
    expect(result.body).toContain('2026-12-31');
  });

  it('defaults documentName to "a document" when not provided', () => {
    const result = emailTemplates.documentExpiring({ name: 'Mia' });
    expect(result.subject).toContain('a document');
    expect(result.body).toContain('a document');
  });

  it('defaults expiresOn to empty string when not provided', () => {
    // Should not throw; body just has an empty date placeholder.
    const result = emailTemplates.documentExpiring({ name: 'Ned' });
    expect(result.body).toBeDefined();
  });

  it('returns both subject and body fields', () => {
    const result = emailTemplates.documentExpiring({ name: 'Olive' });
    expect(result).toHaveProperty('subject');
    expect(result).toHaveProperty('body');
  });

  it('subject contains "expires soon"', () => {
    const result = emailTemplates.documentExpiring({ name: 'Paul', documentName: 'NDA' });
    expect(result.subject).toMatch(/expires soon/i);
  });
});
