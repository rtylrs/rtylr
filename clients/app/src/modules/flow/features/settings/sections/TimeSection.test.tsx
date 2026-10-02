import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

import { initTestI18n } from '../../../shared/testing/i18n';

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

beforeEach(async () => {
  await initTestI18n();
  localStorage.clear();
  useTimeSettingsStore.setState({ ...DEFAULT_TIME_SETTINGS });
});

describe('TimeSection: week and targets', () => {
  it('writes the week start through to the store', () => {
    render(<TimeSection />);

    fireEvent.change(screen.getByLabelText('Week starts'), {
      target: { value: 'sunday' },
    });

    expect(useTimeSettingsStore.getState().weekStart).toBe('sunday');
  });

  it('writes the rounding increment through as a number, not a string', () => {
    render(<TimeSection />);

    fireEvent.change(screen.getByLabelText('Rounding'), {
      target: { value: '30' },
    });

    expect(useTimeSettingsStore.getState().roundingMinutes).toBe(30);
  });

  it('stores one day’s target in minutes while accepting hours, without touching the others', () => {
    render(<TimeSection />);

    fireEvent.change(screen.getByLabelText('Saturday'), {
      target: { value: '6' },
    });

    const { dailyTargetMinutes } = useTimeSettingsStore.getState();
    expect(dailyTargetMinutes.sat).toBe(6 * 60);
    expect(dailyTargetMinutes.mon).toBe(8 * 60);
  });

  it('accepts a zero target, for a genuinely non-working day', () => {
    render(<TimeSection />);

    fireEvent.change(screen.getByLabelText('Friday'), {
      target: { value: '0' },
    });

    expect(useTimeSettingsStore.getState().dailyTargetMinutes.fri).toBe(0);
  });

  it('refuses an overtime threshold below the weekly target', () => {
    render(<TimeSection />);

    fireEvent.change(screen.getByLabelText('Overtime warns past, in hours'), {
      target: { value: '20' },
    });

    expect(useTimeSettingsStore.getState().overtimeThresholdMinutes).toBe(45 * 60);
    expect(screen.getByText(/overtime.*weekly target/i)).toBeInTheDocument();
  });
});

describe('TimeSection: attendance tracking', () => {
  it('defaults to on, preserving current My Day/reports behavior', () => {
    render(<TimeSection />);

    expect(useTimeSettingsStore.getState().trackAttendance).toBe(true);
    expect(
      screen.getByRole('switch', { name: 'Track attendance (Clocked hours)' })
    ).toHaveAttribute('aria-checked', 'true');
  });

  it('writes the toggle through to the store', () => {
    render(<TimeSection />);

    fireEvent.click(screen.getByRole('switch', { name: 'Track attendance (Clocked hours)' }));

    expect(useTimeSettingsStore.getState().trackAttendance).toBe(false);
  });

  it('explains what changes on My Day and reports in each direction', () => {
    render(<TimeSection />);

    expect(screen.getByText(/When on, My Day asks/)).toBeInTheDocument();
  });
});

describe('TimeSection: the missing endpoint', () => {
  it('says plainly that these values are per-browser until the endpoint lands', () => {
    render(<TimeSection />);

    expect(screen.getByText(/These settings are saved in this browser only/)).toBeInTheDocument();
  });
});
