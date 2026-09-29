import { describe, it, expect } from 'vitest';
import { mapRoster, extractRoster } from './rosterWorkflow';

describe('extractRoster', () => {
  it('reads the roster envelope', () => {
    expect(extractRoster({ roster: [{ user_id: 1, name: 'A' }] })).toHaveLength(1);
  });
  it('accepts a bare array', () => {
    expect(extractRoster([{ user_id: 1, name: 'A' }])).toHaveLength(1);
  });
  it('returns [] for unusable input', () => {
    expect(extractRoster(null)).toEqual([]);
    expect(extractRoster({})).toEqual([]);
    expect(extractRoster('x')).toEqual([]);
  });
});

describe('mapRoster', () => {
  it('maps snake_case roster rows', () => {
    const out = mapRoster({
      roster: [
        {
          user_id: 42,
          name: 'Alice',
          has_pin: true,
          can_manage_pos_users: true,
          pos_role: 'manager',
        },
      ],
    });
    expect(out).toEqual([
      { userId: 42, name: 'Alice', hasPin: true, isManager: true, posRole: 'manager' },
    ]);
  });

  it('maps camelCase roster rows', () => {
    const out = mapRoster([{ userId: 7, name: 'Bob', hasPin: false, isManager: false, posRole: '' }]);
    expect(out[0]).toEqual({ userId: 7, name: 'Bob', hasPin: false, isManager: false, posRole: '' });
  });

  it('drops rows without a usable id or name', () => {
    const out = mapRoster({
      roster: [
        { user_id: 0, name: 'Bad' },
        { user_id: 5, name: '   ' },
        { user_id: 5, name: 'Good' },
      ],
    });
    expect(out).toEqual([{ userId: 5, name: 'Good', hasPin: false, isManager: false, posRole: '' }]);
  });

  it('treats a member with no PIN as needing setup', () => {
    const out = mapRoster({ roster: [{ user_id: 9, name: 'New', has_pin: false }] });
    expect(out[0].hasPin).toBe(false);
  });
});
