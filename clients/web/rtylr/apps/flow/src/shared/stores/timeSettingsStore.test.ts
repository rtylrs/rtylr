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
    expect(s.approvalLocksEntry).toBe(true);
    expect(s.autoLockDays).toBe(7);
  });

  it('seeds four categories as a starting point, not a taxonomy', () => {
    const names = useTimeSettingsStore.getState().categories.map((c) => c.name);

    expect(names).toEqual(['Development', 'Review', 'Meeting', 'Admin']);
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

describe('categories', () => {
  it('adds a category with a billable flag', () => {
    useTimeSettingsStore.getState().addCategory('Travel', false);

    const added = useTimeSettingsStore.getState().categories.at(-1);
    expect(added?.name).toBe('Travel');
    expect(added?.billable).toBe(false);
  });

  it('refuses a blank name rather than adding an unnameable row', () => {
    const before = useTimeSettingsStore.getState().categories.length;
    useTimeSettingsStore.getState().addCategory('   ', true);

    expect(useTimeSettingsStore.getState().categories).toHaveLength(before);
  });

  it('refuses a duplicate name, case-insensitively', () => {
    const before = useTimeSettingsStore.getState().categories.length;
    useTimeSettingsStore.getState().addCategory('development', true);

    expect(useTimeSettingsStore.getState().categories).toHaveLength(before);
  });

  it('renames a category in place', () => {
    const id = useTimeSettingsStore.getState().categories[0].id;
    useTimeSettingsStore.getState().renameCategory(id, 'Engineering');

    expect(useTimeSettingsStore.getState().categories[0].name).toBe('Engineering');
  });

  it('toggles the billable flag', () => {
    const first = useTimeSettingsStore.getState().categories[0];
    useTimeSettingsStore.getState().setCategoryBillable(first.id, !first.billable);

    expect(useTimeSettingsStore.getState().categories[0].billable).toBe(!first.billable);
  });

  it('deletes a category', () => {
    const id = useTimeSettingsStore.getState().categories[0].id;
    useTimeSettingsStore.getState().deleteCategory(id);

    expect(useTimeSettingsStore.getState().categories.map((c) => c.id)).not.toContain(id);
  });

  it('keeps the last category, because an entry with no category to pick is unloggable', () => {
    const { categories } = useTimeSettingsStore.getState();
    categories.slice(1).forEach((c) => useTimeSettingsStore.getState().deleteCategory(c.id));
    const lastId = useTimeSettingsStore.getState().categories[0].id;

    useTimeSettingsStore.getState().deleteCategory(lastId);

    expect(useTimeSettingsStore.getState().categories).toHaveLength(1);
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
