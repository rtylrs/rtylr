import { describe, it, expect, beforeEach } from 'vitest';
import {
  DEFAULT_TIME_SETTINGS,
  TIME_SETTINGS_STORAGE_KEY,
  getWeeklyTargetMinutes,
  migrateTimeSettingsV0,
  migrateTimeSettingsV1,
  roundLoggedMinutes,
  useTimeSettingsStore,
} from './timeSettingsStore';

beforeEach(() => {
  localStorage.clear();
  useTimeSettingsStore.setState({ ...DEFAULT_TIME_SETTINGS });
});

describe('time settings defaults', () => {
  it('seeds the shipped defaults, so no screen has to hardcode them', () => {
    const s = useTimeSettingsStore.getState();

    expect(s.weekStart).toBe('monday');
    expect(s.roundingMinutes).toBe(15);
    expect(s.dailyTargetMinutes).toEqual({
      sun: 0,
      mon: 8 * 60,
      tue: 8 * 60,
      wed: 8 * 60,
      thu: 8 * 60,
      fri: 8 * 60,
      sat: 4 * 60,
    });
    expect(getWeeklyTargetMinutes(s.dailyTargetMinutes)).toBe(44 * 60);
    expect(s.overtimeThresholdMinutes).toBe(45 * 60);
  });
});

describe('daily targets', () => {
  it('writes one day without touching the others', () => {
    useTimeSettingsStore.getState().setDailyTargetMinutes('sat', 6 * 60);

    const { dailyTargetMinutes } = useTimeSettingsStore.getState();
    expect(dailyTargetMinutes.sat).toBe(6 * 60);
    expect(dailyTargetMinutes.mon).toBe(8 * 60);
    expect(dailyTargetMinutes.sun).toBe(0);
  });

  it('allows a day to be set to zero, for a genuinely non-working day', () => {
    useTimeSettingsStore.getState().setDailyTargetMinutes('fri', 0);

    expect(useTimeSettingsStore.getState().dailyTargetMinutes.fri).toBe(0);
  });

  it('sums every day for the weekly total', () => {
    useTimeSettingsStore.getState().setDailyTargetMinutes('sat', 0);

    expect(getWeeklyTargetMinutes(useTimeSettingsStore.getState().dailyTargetMinutes)).toBe(
      5 * 8 * 60
    );
  });
});

describe('migrateTimeSettingsV0', () => {
  it('spreads a flat weekly target evenly across Mon-Fri, leaving weekends at zero', () => {
    const migrated = migrateTimeSettingsV0({ weeklyTargetMinutes: 35 * 60 }) as {
      dailyTargetMinutes: Record<string, number>;
    };

    expect(migrated.dailyTargetMinutes).toEqual({
      sun: 0,
      mon: 7 * 60,
      tue: 7 * 60,
      wed: 7 * 60,
      thu: 7 * 60,
      fri: 7 * 60,
      sat: 0,
    });
  });

  it('falls back to the 40h default when the old field is missing', () => {
    const migrated = migrateTimeSettingsV0({}) as { dailyTargetMinutes: Record<string, number> };

    expect(migrated.dailyTargetMinutes.mon).toBe(8 * 60);
  });

  it('preserves every other field untouched', () => {
    const migrated = migrateTimeSettingsV0({
      weeklyTargetMinutes: 40 * 60,
      weekStart: 'sunday',
    }) as { weekStart: string };

    expect(migrated.weekStart).toBe('sunday');
  });
});

describe('migrateTimeSettingsV1', () => {
  it('drops a stored rate table, so pay rates do not linger in a browser', () => {
    const stored = {
      weekStart: 'monday',
      ratesByUserId: { '7': 95, '9': 120 },
      roundingMinutes: 15,
    };

    const migrated = migrateTimeSettingsV1(stored) as Record<string, unknown>;

    expect('ratesByUserId' in migrated).toBe(false);
    // Everything else survives: this is a targeted removal, not a reset. A
    // workspace that customised its week start keeps it.
    expect(migrated.weekStart).toBe('monday');
    expect(migrated.roundingMinutes).toBe(15);
  });

  it('is a no-op on a blob that never had rates', () => {
    const migrated = migrateTimeSettingsV1({ weekStart: 'sunday' }) as Record<string, unknown>;

    expect(migrated).toEqual({ weekStart: 'sunday' });
  });
});

describe('persistence', () => {
  it('writes through to storage under a versioned key', () => {
    useTimeSettingsStore.getState().setWeekStart('sunday');

    const raw = localStorage.getItem(TIME_SETTINGS_STORAGE_KEY);
    expect(raw).toContain('sunday');
  });
});

describe('roundLoggedMinutes', () => {
  it('leaves a duration alone when rounding is off', () => {
    expect(roundLoggedMinutes(7, 0)).toBe(7);
  });

  it('rounds to the nearest increment', () => {
    expect(roundLoggedMinutes(100, 15)).toBe(105);
    expect(roundLoggedMinutes(97, 15)).toBe(90);
  });

  it('never rounds a real entry down to nothing', () => {
    expect(roundLoggedMinutes(5, 15)).toBe(15);
    expect(roundLoggedMinutes(1, 30)).toBe(30);
  });

  it('leaves an exact multiple untouched', () => {
    expect(roundLoggedMinutes(105, 15)).toBe(105);
  });
});

describe('retired time configuration', () => {
  it('rehydrates working preferences and removes unused v2 fields', async () => {
    localStorage.setItem(
      TIME_SETTINGS_STORAGE_KEY,
      JSON.stringify({
        version: 2,
        state: {
          weekStart: 'sunday',
          roundingMinutes: 30,
          approvalLocksEntry: false,
          autoLockDays: 3,
          categories: [{ id: 'legacy', name: 'Legacy', billable: true }],
        },
      })
    );
    await useTimeSettingsStore.persist.rehydrate();
    expect(useTimeSettingsStore.getState().weekStart).toBe('sunday');
    expect(useTimeSettingsStore.getState().roundingMinutes).toBe(30);
    const saved = JSON.parse(localStorage.getItem(TIME_SETTINGS_STORAGE_KEY)!);
    expect(saved.state).not.toHaveProperty('approvalLocksEntry');
    expect(saved.state).not.toHaveProperty('autoLockDays');
    expect(saved.state).not.toHaveProperty('categories');
  });
});
