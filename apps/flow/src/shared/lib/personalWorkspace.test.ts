import { describe, expect, it } from 'vitest';
import { isApiWorkspaceId, PERSONAL_WORKSPACE_ID } from './personalWorkspace';

describe('isApiWorkspaceId', () => {
  it('rejects a missing id', () => {
    expect(isApiWorkspaceId(undefined)).toBe(false);
    expect(isApiWorkspaceId('')).toBe(false);
  });

  it('rejects the client-only Personal sentinel', () => {
    expect(isApiWorkspaceId(PERSONAL_WORKSPACE_ID)).toBe(false);
  });

  it('accepts a resolved numeric workspace id', () => {
    expect(isApiWorkspaceId('1')).toBe(true);
    expect(isApiWorkspaceId('4821')).toBe(true);
  });

  it('rejects ids the backend @gt(0) constraint would reject', () => {
    expect(isApiWorkspaceId('0')).toBe(false);
    expect(isApiWorkspaceId('-3')).toBe(false);
  });

  it('rejects ids that are not whole numbers', () => {
    expect(isApiWorkspaceId('1.5')).toBe(false);
    expect(isApiWorkspaceId('12abc')).toBe(false);
    expect(isApiWorkspaceId(' ')).toBe(false);
  });
});
