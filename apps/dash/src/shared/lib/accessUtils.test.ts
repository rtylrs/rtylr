import { describe, it, expect } from 'vitest';
import {
  hasAccess,
  mergeAccess,
  accessLabel,
  isAdmin,
  canWrite,
  canRead,
  resolveAccess,
  allAccessLevels,
} from './accessUtils';

describe('hasAccess', () => {
  it('returns true when level meets required', () => {
    expect(hasAccess('admin', 'write')).toBe(true);
    expect(hasAccess('write', 'write')).toBe(true);
    expect(hasAccess('read', 'read')).toBe(true);
  });

  it('returns false when level is below required', () => {
    expect(hasAccess('read', 'write')).toBe(false);
    expect(hasAccess('none', 'read')).toBe(false);
  });
});

describe('mergeAccess', () => {
  it('returns higher access', () => {
    expect(mergeAccess('read', 'write')).toBe('write');
    expect(mergeAccess('admin', 'none')).toBe('admin');
    expect(mergeAccess('none', 'none')).toBe('none');
  });
});

describe('accessLabel', () => {
  it('returns correct labels', () => {
    expect(accessLabel('none')).toBe('No Access');
    expect(accessLabel('read')).toBe('Read Only');
    expect(accessLabel('write')).toBe('Read & Write');
    expect(accessLabel('admin')).toBe('Administrator');
  });
});

describe('isAdmin', () => {
  it('true for admin', () => expect(isAdmin('admin')).toBe(true));
  it('false for others', () => expect(isAdmin('write')).toBe(false));
});

describe('canWrite', () => {
  it('true for write and admin', () => {
    expect(canWrite('write')).toBe(true);
    expect(canWrite('admin')).toBe(true);
  });

  it('false for read and none', () => {
    expect(canWrite('read')).toBe(false);
    expect(canWrite('none')).toBe(false);
  });
});

describe('canRead', () => {
  it('true for read, write, admin', () => {
    expect(canRead('read')).toBe(true);
    expect(canRead('write')).toBe(true);
    expect(canRead('admin')).toBe(true);
  });

  it('false for none', () => expect(canRead('none')).toBe(false));
});

describe('resolveAccess', () => {
  it('uses resource override when provided', () => {
    expect(resolveAccess('admin', 'read')).toBe('read');
    expect(resolveAccess('none', 'write')).toBe('write');
  });

  it('falls back to user level when no override', () => {
    expect(resolveAccess('write')).toBe('write');
    expect(resolveAccess('admin', null)).toBe('admin');
  });
});

describe('allAccessLevels', () => {
  it('returns all four levels', () => {
    expect(allAccessLevels()).toHaveLength(4);
    expect(allAccessLevels()[0]).toBe('none');
    expect(allAccessLevels()[3]).toBe('admin');
  });
});
