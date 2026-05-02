import { describe, it, expect } from 'vitest';

import { mapMeResponseToFlowUser } from './mapMeUser';

describe('mapMeResponseToFlowUser', () => {
  it('returns null for null input', () => {
    expect(mapMeResponseToFlowUser(null)).toBeNull();
  });

  it('returns null for undefined input', () => {
    expect(mapMeResponseToFlowUser(undefined)).toBeNull();
  });

  it('returns null for a non-object primitive', () => {
    expect(
      mapMeResponseToFlowUser('not-an-object' as unknown as Record<string, unknown>)
    ).toBeNull();
  });

  it('maps fullName to name', () => {
    const result = mapMeResponseToFlowUser({ id: '1', email: 'a@a.com', fullName: 'Alice Smith' });
    expect(result?.name).toBe('Alice Smith');
  });

  it('derives name from firstName + lastName when fullName is absent', () => {
    const result = mapMeResponseToFlowUser({
      id: '2',
      email: 'b@b.com',
      firstName: 'Bob',
      lastName: 'Jones',
    });
    expect(result?.name).toBe('Bob Jones');
  });

  it('uses the single name field as fallback', () => {
    const result = mapMeResponseToFlowUser({ id: '3', email: 'c@c.com', name: 'Charlie' });
    expect(result?.name).toBe('Charlie');
  });

  it('falls back to email when no name fields are present', () => {
    const result = mapMeResponseToFlowUser({ id: '4', email: 'nobody@example.com' });
    expect(result?.name).toBe('nobody@example.com');
  });

  it('falls back to "User" when everything is absent', () => {
    const result = mapMeResponseToFlowUser({});
    expect(result?.name).toBe('User');
  });

  it('stringifies numeric id', () => {
    const result = mapMeResponseToFlowUser({
      id: 42 as unknown as string,
      email: 'd@d.com',
      fullName: 'Dave',
    });
    expect(result?.id).toBe('42');
  });

  it('uses userId as id fallback', () => {
    const result = mapMeResponseToFlowUser({ userId: '99', email: 'e@e.com', fullName: 'Eve' });
    expect(result?.id).toBe('99');
  });

  it('trims whitespace from email', () => {
    const result = mapMeResponseToFlowUser({ id: '5', email: '  frank@f.com  ' });
    expect(result?.email).toBe('frank@f.com');
  });

  it('passes avatar through when present', () => {
    const result = mapMeResponseToFlowUser({
      id: '6',
      email: 'g@g.com',
      fullName: 'Grace',
      avatar: 'https://cdn/avatar.png',
    });
    expect(result?.avatar).toBe('https://cdn/avatar.png');
  });

  it('omits avatar when not a string', () => {
    const result = mapMeResponseToFlowUser({
      id: '7',
      email: 'h@h.com',
      fullName: 'Hank',
      avatar: 42,
    });
    expect(result?.avatar).toBeUndefined();
  });
});
