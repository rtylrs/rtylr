import { describe, it, expect } from 'vitest';
import { templateSchema } from '../templates';

const baseTemplate = {
  id: '1',
  name: 'Welcome Email',
  channel: 'email',
  category: 'onboarding',
  body: 'Hello {{name}}, welcome!',
  updatedAt: '2024-04-01',
};

describe('templateSchema', () => {
  it('parses a valid email template', () => {
    const result = templateSchema.parse(baseTemplate);
    expect(result.id).toBe('1');
    expect(result.name).toBe('Welcome Email');
    expect(result.channel).toBe('email');
  });

  it('parses a valid sms template', () => {
    const result = templateSchema.parse({ ...baseTemplate, channel: 'sms', id: '2' });
    expect(result.channel).toBe('sms');
  });

  it('falls back to email for unknown channel', () => {
    const result = templateSchema.parse({ ...baseTemplate, channel: 'fax' });
    expect(result.channel).toBe('email');
  });

  it('keeps id as a string', () => {
    const result = templateSchema.parse({ ...baseTemplate, id: '7' });
    expect(result.id).toBe('7');
  });

  it('allows optional subject', () => {
    const result = templateSchema.parse({ ...baseTemplate, subject: 'Welcome to our club!' });
    expect(result.subject).toBe('Welcome to our club!');
  });

  it('allows null subject', () => {
    const result = templateSchema.parse({ ...baseTemplate, subject: null });
    expect(result.subject).toBeNull();
  });

  it('allows null thumbnail', () => {
    const result = templateSchema.parse({ ...baseTemplate, thumbnail: null });
    expect(result.thumbnail).toBeNull();
  });

  it('stores the body correctly', () => {
    const result = templateSchema.parse(baseTemplate);
    expect(result.body).toBe('Hello {{name}}, welcome!');
  });
});
