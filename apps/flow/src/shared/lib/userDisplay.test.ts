import { describe, it, expect } from 'vitest';

import { displayUserEmail, displayUserName } from './userDisplay';
import type { FlowUser } from '../stores/userStore';

function makeUser(overrides: Partial<FlowUser> = {}): FlowUser {
  return {
    id: 'u1',
    name: '',
    email: '',
    ...overrides,
  };
}

// ---------------------------------------------------------------------------
// displayUserEmail
// ---------------------------------------------------------------------------

describe('displayUserEmail', () => {
  it('returns empty string for null', () => {
    expect(displayUserEmail(null)).toBe('');
  });

  it('returns empty string for undefined', () => {
    expect(displayUserEmail(undefined)).toBe('');
  });

  it('trims surrounding whitespace', () => {
    expect(displayUserEmail(makeUser({ email: '  alice@example.com  ' }))).toBe(
      'alice@example.com'
    );
  });

  it('returns the email when present', () => {
    expect(displayUserEmail(makeUser({ email: 'bob@example.com' }))).toBe('bob@example.com');
  });
});

// ---------------------------------------------------------------------------
// displayUserName
// ---------------------------------------------------------------------------

describe('displayUserName', () => {
  it('returns empty string for null', () => {
    expect(displayUserName(null)).toBe('');
  });

  it('returns empty string for undefined', () => {
    expect(displayUserName(undefined)).toBe('');
  });

  it('prefers name over firstName and email', () => {
    expect(
      displayUserName(makeUser({ name: 'Alice Smith', firstName: 'Alice', email: 'a@a.com' }))
    ).toBe('Alice Smith');
  });

  it('falls back to firstName when name is empty', () => {
    expect(displayUserName(makeUser({ name: '', firstName: 'Alice', email: 'a@a.com' }))).toBe(
      'Alice'
    );
  });

  it('falls back to email when name and firstName are empty', () => {
    expect(displayUserName(makeUser({ name: '', firstName: '', email: 'a@a.com' }))).toBe(
      'a@a.com'
    );
  });

  it('returns empty string when all fields are empty', () => {
    expect(displayUserName(makeUser({ name: '', firstName: '', email: '' }))).toBe('');
  });

  it('trims whitespace from name', () => {
    expect(displayUserName(makeUser({ name: '  Bob  ' }))).toBe('Bob');
  });
});
