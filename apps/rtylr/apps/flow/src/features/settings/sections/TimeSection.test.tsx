import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

// Interpolates params, unlike the bare key-returning mock used elsewhere: the
// labels under test are only distinguishable once their values are substituted.
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, params?: Record<string, unknown>) =>
      params ? `${key}:${Object.values(params).join(',')}` : key,
    i18n: { language: 'en-US' },
  }),
}));

vi.mock('../../../shared/hooks/queries/useFlowTeam', () => ({
  useFlowTeam: () => ({
    data: [
      { id: '7', name: 'Maya Krishnan' },
      { id: '9', name: 'Sam Okafor' },
    ],
    isLoading: false,
  }),
}));

// WorkTypesSection pulls in useWorkDay, which imports the real i18next
// singleton for toasts outside the component tree - that singleton calls
// initReactI18next on import, which the react-i18next mock above doesn't
// export. Mock the hook module itself, same as LogTimeDialog.test.tsx does.
vi.mock('../../time/useWorkDay', () => ({
  useWorkTypes: () => ({ data: [], isLoading: false }),
  useCreateWorkType: () => ({ mutate: vi.fn(), isPending: false }),
  useUpdateWorkType: () => ({ mutate: vi.fn(), isPending: false }),
  useDeleteWorkType: () => ({ mutate: vi.fn(), isPending: false }),
}));

import TimeSection from './TimeSection';
import {
  DEFAULT_TIME_SETTINGS,
  useTimeSettingsStore,
} from '../../../shared/stores/timeSettingsStore';

beforeEach(() => {
  localStorage.clear();
  useTimeSettingsStore.setState({ ...DEFAULT_TIME_SETTINGS });
});

describe('TimeSection: week and targets', () => {
  it('writes the week start through to the store', () => {
    render(<TimeSection />);

    fireEvent.change(screen.getByLabelText('flow.settings.time.weekStart'), {
      target: { value: 'sunday' },
    });

    expect(useTimeSettingsStore.getState().weekStart).toBe('sunday');
  });

  it('writes the rounding increment through as a number, not a string', () => {
    render(<TimeSection />);

    fireEvent.change(screen.getByLabelText('flow.settings.time.rounding'), {
      target: { value: '30' },
    });

    expect(useTimeSettingsStore.getState().roundingMinutes).toBe(30);
  });

  it('stores one day’s target in minutes while accepting hours, without touching the others', () => {
    render(<TimeSection />);

    fireEvent.change(screen.getByLabelText('flow.settings.time.saturday'), {
      target: { value: '6' },
    });

    const { dailyTargetMinutes } = useTimeSettingsStore.getState();
    expect(dailyTargetMinutes.sat).toBe(6 * 60);
    expect(dailyTargetMinutes.mon).toBe(8 * 60);
  });

  it('accepts a zero target, for a genuinely non-working day', () => {
    render(<TimeSection />);

    fireEvent.change(screen.getByLabelText('flow.settings.time.friday'), {
      target: { value: '0' },
    });

    expect(useTimeSettingsStore.getState().dailyTargetMinutes.fri).toBe(0);
  });

  it('refuses an overtime threshold below the weekly target', () => {
    render(<TimeSection />);

    fireEvent.change(screen.getByLabelText('flow.settings.time.overtimeThreshold'), {
      target: { value: '20' },
    });

    expect(useTimeSettingsStore.getState().overtimeThresholdMinutes).toBe(45 * 60);
    expect(screen.getByText('flow.settings.time.overtimeBelowTarget')).toBeInTheDocument();
  });
});

describe('TimeSection: the approval lock', () => {
  it('states both irreversible consequences before the switch is touched', () => {
    render(<TimeSection />);

    expect(screen.getByText('flow.settings.time.lockConsequence')).toBeInTheDocument();
  });

  it('offers an undo after flipping it, because it changes what everyone else can edit', () => {
    render(<TimeSection />);

    fireEvent.click(screen.getByRole('switch', { name: 'flow.settings.time.approvalLocks' }));

    expect(useTimeSettingsStore.getState().approvalLocksEntry).toBe(false);
    fireEvent.click(screen.getByRole('button', { name: 'flow.settings.time.undo' }));
    expect(useTimeSettingsStore.getState().approvalLocksEntry).toBe(true);
  });
});

describe('TimeSection: categories', () => {
  it('adds a category', () => {
    render(<TimeSection />);

    fireEvent.change(screen.getByLabelText('flow.settings.time.newCategoryName'), {
      target: { value: 'Travel' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'flow.settings.time.addCategory' }));

    expect(useTimeSettingsStore.getState().categories.map((c) => c.name)).toContain('Travel');
  });

  it('labels each remove control by the category it removes, not by the verb', () => {
    render(<TimeSection />);

    expect(
      screen.getByRole('button', { name: 'flow.settings.time.removeCategoryLabel:Development' })
    ).toBeInTheDocument();
    expect(
      screen.getByRole('button', { name: 'flow.settings.time.removeCategoryLabel:Admin' })
    ).toBeInTheDocument();
  });

  it('says a category configures nothing yet, because TimeLog has no category field', () => {
    render(<TimeSection />);

    expect(screen.getByText('flow.settings.time.categoriesNotWired')).toBeInTheDocument();
  });
});

describe('TimeSection: attendance tracking', () => {
  it('defaults to on, preserving current My Day/reports behavior', () => {
    render(<TimeSection />);

    expect(useTimeSettingsStore.getState().trackAttendance).toBe(true);
    expect(
      screen.getByRole('switch', { name: 'flow.settings.time.trackAttendance' })
    ).toHaveAttribute('aria-checked', 'true');
  });

  it('writes the toggle through to the store', () => {
    render(<TimeSection />);

    fireEvent.click(screen.getByRole('switch', { name: 'flow.settings.time.trackAttendance' }));

    expect(useTimeSettingsStore.getState().trackAttendance).toBe(false);
  });

  it('explains what changes on My Day and reports in each direction', () => {
    render(<TimeSection />);

    expect(screen.getByText('flow.settings.time.trackAttendanceExplained')).toBeInTheDocument();
  });
});

describe('TimeSection: the missing endpoint', () => {
  it('says plainly that these values are per-browser until the endpoint lands', () => {
    render(<TimeSection />);

    expect(screen.getByText('flow.settings.time.localOnlyWarning')).toBeInTheDocument();
  });
});
