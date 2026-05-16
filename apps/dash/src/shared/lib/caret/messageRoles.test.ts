import { describe, expect, it } from 'vitest';
import { roleLabel } from './messageRoles';

describe('roleLabel', () => {
  it('returns "You" for user', () => {
    expect(roleLabel('user')).toBe('You');
  });
  it('returns "Caret" for assistant', () => {
    expect(roleLabel('assistant')).toBe('Caret');
  });
  it('returns "Tool" for tool', () => {
    expect(roleLabel('tool')).toBe('Tool');
  });
  it('returns "System" for system', () => {
    expect(roleLabel('system')).toBe('System');
  });
});
