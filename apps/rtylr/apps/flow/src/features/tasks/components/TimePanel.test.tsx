import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { ApiTimeLog } from '../../../shared/services/flowApi';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key, i18n: { language: 'en-US' } }),
}));

const logTimeMutate = vi.hoisted(() => vi.fn());
const deleteTimeLogMutate = vi.hoisted(() => vi.fn());
const refetch = vi.hoisted(() => vi.fn());

const timeLogsState = vi.hoisted(() => ({
  value: {
    data: undefined as { timeLogs: ApiTimeLog[] } | undefined,
    isLoading: false,
    isError: false,
    refetch,
  },
}));
const logTimeState = vi.hoisted(() => ({
  value: { mutate: logTimeMutate, isPending: false, variables: undefined as unknown },
}));

vi.mock('../useTimeLogs', () => ({
  useTimeLogs: () => timeLogsState.value,
  useLogTime: () => logTimeState.value,
  useDeleteTimeLog: () => ({ mutate: deleteTimeLogMutate, isPending: false }),
}));

vi.mock('../../../shared/hooks/queries/useFlowTeam', () => ({
  useFlowTeam: () => ({
    data: [
      { id: '7', name: 'Maya Krishnan' },
      { id: '9', name: 'Sam Okafor' },
    ],
  }),
}));

import TimePanel from './TimePanel';
import {
  DEFAULT_TIME_SETTINGS,
  useTimeSettingsStore,
} from '../../../shared/stores/timeSettingsStore';

function makeLog(over: Partial<ApiTimeLog> = {}): ApiTimeLog {
  return {
    id: 1,
    taskId: 42,
    userId: 7,
    durationMinutes: 195,
    description: 'Rewrote the migration and reran it against staging',
    loggedAt: '2026-08-11T09:00:00Z',
    createdAt: '2026-08-11T09:00:00Z',
    ...over,
  };
}

function renderPanel(props: Partial<React.ComponentProps<typeof TimePanel>> = {}) {
  return render(
    <TimePanel
      taskId="42"
      estimateHours={8}
      canLog
      canLogForOthers={false}
      currentUserId="7"
      {...props}
    />
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  timeLogsState.value = { data: undefined, isLoading: false, isError: false, refetch };
  logTimeState.value = { mutate: logTimeMutate, isPending: false, variables: undefined };
  useTimeSettingsStore.setState({ ...DEFAULT_TIME_SETTINGS });
});

describe('TimePanel: loading', () => {
  it('holds the panel height with shimmer rows rather than a spinner that collapses it', () => {
    timeLogsState.value = { data: undefined, isLoading: true, isError: false, refetch };

    const { container } = renderPanel();

    expect(container.querySelectorAll('.flow-shimmer')).toHaveLength(2);
    expect(screen.queryByRole('status', { name: /loading/i })).not.toBeInTheDocument();
  });
});

describe('TimePanel: empty', () => {
  it('states there is nothing logged and offers the first entry', () => {
    timeLogsState.value = { data: { timeLogs: [] }, isLoading: false, isError: false, refetch };

    renderPanel();

    expect(screen.getByText('flow.time.empty')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'flow.time.logFirst' })).toBeInTheDocument();
  });

  it('hides the estimate bar entirely rather than drawing it at zero', () => {
    timeLogsState.value = { data: { timeLogs: [] }, isLoading: false, isError: false, refetch };

    renderPanel();

    expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
  });

  it('offers a refresh instead of the log form to a viewer who cannot log', () => {
    timeLogsState.value = { data: { timeLogs: [] }, isLoading: false, isError: false, refetch };

    renderPanel({ canLog: false });

    expect(screen.queryByRole('button', { name: 'flow.time.logFirst' })).not.toBeInTheDocument();
    const retry = screen.getByRole('button', { name: 'flow.time.retry' });
    fireEvent.click(retry);
    expect(refetch).toHaveBeenCalled();
  });
});

describe('TimePanel: read error', () => {
  it('says the read failed and offers a retry', () => {
    timeLogsState.value = { data: undefined, isLoading: false, isError: true, refetch };

    renderPanel();

    expect(screen.getByText('flow.time.loadFailed')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'flow.time.retry' }));
    expect(refetch).toHaveBeenCalledOnce();
  });

  it('leaves the write path live, because reading and writing are separate endpoints', () => {
    timeLogsState.value = { data: undefined, isLoading: false, isError: true, refetch };

    renderPanel();
    fireEvent.click(screen.getByRole('button', { name: 'flow.time.logTime' }));

    expect(screen.getByLabelText('flow.time.duration')).toBeEnabled();
  });
});

describe('TimePanel: entries', () => {
  it('totals logged minutes against the estimate in a padded clock figure', () => {
    timeLogsState.value = {
      data: {
        timeLogs: [
          makeLog({ id: 1, durationMinutes: 195 }),
          makeLog({ id: 2, durationMinutes: 60 }),
        ],
      },
      isLoading: false,
      isError: false,
      refetch,
    };

    renderPanel();

    expect(screen.getByText('4:15')).toBeInTheDocument();
    expect(screen.getByText('8:00')).toBeInTheDocument();
  });

  it('announces the running total, so a screen reader hears the new figure', () => {
    // Two entries, so the total differs from either row. With rows now in the
    // same clock notation, a single entry makes the total ambiguous.
    timeLogsState.value = {
      data: {
        timeLogs: [
          makeLog({ id: 1, durationMinutes: 195 }),
          makeLog({ id: 2, durationMinutes: 60 }),
        ],
      },
      isLoading: false,
      isError: false,
      refetch,
    };

    const { container } = renderPanel();

    expect(container.querySelector('[aria-live="polite"]')).toHaveTextContent('4:15');
  });

  it('names the author from the team rather than showing a raw user id', () => {
    timeLogsState.value = {
      data: { timeLogs: [makeLog({ userId: 9 })] },
      isLoading: false,
      isError: false,
      refetch,
    };

    renderPanel();

    expect(screen.getByText('Sam Okafor')).toBeInTheDocument();
  });

  it('renders entry durations in clock notation, so the column aligns', () => {
    timeLogsState.value = {
      data: {
        timeLogs: [
          makeLog({ id: 1, durationMinutes: 195 }),
          makeLog({ id: 2, durationMinutes: 60 }),
        ],
      },
      isLoading: false,
      isError: false,
      refetch,
    };

    renderPanel();

    expect(screen.getByText('3:15')).toBeInTheDocument();
    expect(screen.getByText('1:00')).toBeInTheDocument();
    expect(screen.queryByText('3h 15m')).not.toBeInTheDocument();
  });

  it('offers delete on your own entry, labelled by what it removes', () => {
    timeLogsState.value = {
      data: { timeLogs: [makeLog({ id: 5, userId: 7 })] },
      isLoading: false,
      isError: false,
      refetch,
    };

    renderPanel();
    fireEvent.click(screen.getByRole('button', { name: 'flow.time.deleteEntryLabel' }));

    expect(deleteTimeLogMutate).toHaveBeenCalledWith({ id: 5, taskId: '42' });
  });

  it('does not offer delete to a viewer, even on their own entry', () => {
    timeLogsState.value = {
      data: { timeLogs: [makeLog({ id: 5, userId: 7 })] },
      isLoading: false,
      isError: false,
      refetch,
    };

    renderPanel({ canLog: false });

    expect(
      screen.queryByRole('button', { name: 'flow.time.deleteEntryLabel' })
    ).not.toBeInTheDocument();
  });

  it('does not offer delete on someone else’s entry', () => {
    timeLogsState.value = {
      data: { timeLogs: [makeLog({ id: 5, userId: 9 })] },
      isLoading: false,
      isError: false,
      refetch,
    };

    renderPanel();

    expect(
      screen.queryByRole('button', { name: 'flow.time.deleteEntryLabel' })
    ).not.toBeInTheDocument();
  });
});

describe('TimePanel: optimistic save', () => {
  it('shows the in-flight entry before the server confirms it', () => {
    timeLogsState.value = { data: { timeLogs: [] }, isLoading: false, isError: false, refetch };
    logTimeState.value = {
      mutate: logTimeMutate,
      isPending: true,
      variables: {
        task_id: 42,
        duration_minutes: 105,
        description: 'Pairing on the rollback path',
      },
    };

    renderPanel();

    expect(screen.getByText('1:45')).toBeInTheDocument();
    expect(screen.getByText('flow.time.saving')).toBeInTheDocument();
  });
});

describe('TimePanel: the log form', () => {
  beforeEach(() => {
    timeLogsState.value = { data: { timeLogs: [] }, isLoading: false, isError: false, refetch };
  });

  function openForm() {
    renderPanel();
    fireEvent.click(screen.getByRole('button', { name: 'flow.time.logFirst' }));
  }

  it('fixes the date to today and states why, because the create RPC takes no date', () => {
    openForm();

    const date = screen.getByLabelText('flow.time.date');
    expect(date).toBeDisabled();
    const describedBy = date.getAttribute('aria-describedby') as string;
    expect(document.getElementById(describedBy)).toHaveTextContent('flow.time.dateFixedToToday');
  });

  it('says you cannot log against someone else when that permission is missing', () => {
    openForm();

    expect(screen.getByText('flow.time.logForOthersGate')).toBeInTheDocument();
  });

  it('omits the log-for-others reason when the permission is held', () => {
    renderPanel({ canLogForOthers: true });
    fireEvent.click(screen.getByRole('button', { name: 'flow.time.logFirst' }));

    expect(screen.queryByText('flow.time.logForOthersGate')).not.toBeInTheDocument();
  });

  it('refuses to submit a duration it cannot parse', () => {
    openForm();
    fireEvent.change(screen.getByLabelText('flow.time.duration'), { target: { value: 'soon' } });
    fireEvent.keyDown(screen.getByLabelText('flow.time.duration'), { key: 'Enter' });

    expect(logTimeMutate).not.toHaveBeenCalled();
    expect(screen.getByLabelText('flow.time.duration')).toHaveAttribute('aria-invalid', 'true');
  });

  it('submits parsed minutes on Enter', () => {
    openForm();
    fireEvent.change(screen.getByLabelText('flow.time.duration'), { target: { value: '1:45' } });
    fireEvent.keyDown(screen.getByLabelText('flow.time.duration'), { key: 'Enter' });

    expect(logTimeMutate).toHaveBeenCalledWith(
      { task_id: 42, duration_minutes: 105, description: undefined },
      expect.anything()
    );
  });

  it('closes on Escape and returns focus to the button that opened it', () => {
    openForm();
    fireEvent.keyDown(screen.getByLabelText('flow.time.duration'), { key: 'Escape' });

    const trigger = screen.getByRole('button', { name: 'flow.time.logFirst' });
    expect(screen.queryByLabelText('flow.time.duration')).not.toBeInTheDocument();
    expect(trigger).toHaveFocus();
  });

  it('is absent for a viewer who cannot log at all', () => {
    renderPanel({ canLog: false });

    expect(screen.queryByRole('button', { name: 'flow.time.logFirst' })).not.toBeInTheDocument();
  });
});

describe('TimePanel: workspace rounding', () => {
  beforeEach(() => {
    timeLogsState.value = { data: { timeLogs: [] }, isLoading: false, isError: false, refetch };
    useTimeSettingsStore.setState({ roundingMinutes: 15 });
  });

  it('applies the workspace rounding increment to what it posts', () => {
    renderPanel();
    fireEvent.click(screen.getByRole('button', { name: 'flow.time.logFirst' }));
    fireEvent.change(screen.getByLabelText('flow.time.duration'), { target: { value: '52m' } });
    fireEvent.keyDown(screen.getByLabelText('flow.time.duration'), { key: 'Enter' });

    expect(logTimeMutate).toHaveBeenCalledWith(
      expect.objectContaining({ duration_minutes: 45 }),
      expect.anything()
    );
  });

  it('shows the rounded figure on the submit button rather than rounding silently', () => {
    renderPanel();
    fireEvent.click(screen.getByRole('button', { name: 'flow.time.logFirst' }));
    fireEvent.change(screen.getByLabelText('flow.time.duration'), { target: { value: '52m' } });

    expect(screen.getByRole('button', { name: /flow\.time\.logAmount/ })).toBeInTheDocument();
    expect(screen.getByText('flow.time.roundedNotice')).toBeInTheDocument();
  });

  it('says nothing about rounding when the value is already an exact increment', () => {
    renderPanel();
    fireEvent.click(screen.getByRole('button', { name: 'flow.time.logFirst' }));
    fireEvent.change(screen.getByLabelText('flow.time.duration'), { target: { value: '45m' } });

    expect(screen.queryByText('flow.time.roundedNotice')).not.toBeInTheDocument();
  });

  it('posts the exact duration when rounding is switched off', () => {
    useTimeSettingsStore.setState({ roundingMinutes: 0 });
    renderPanel();
    fireEvent.click(screen.getByRole('button', { name: 'flow.time.logFirst' }));
    fireEvent.change(screen.getByLabelText('flow.time.duration'), { target: { value: '52m' } });
    fireEvent.keyDown(screen.getByLabelText('flow.time.duration'), { key: 'Enter' });

    expect(logTimeMutate).toHaveBeenCalledWith(
      expect.objectContaining({ duration_minutes: 52 }),
      expect.anything()
    );
  });
});

describe('TimePanel: over estimate', () => {
  it('flags the overrun as one pill and caps the bar at full width', () => {
    timeLogsState.value = {
      data: { timeLogs: [makeLog({ durationMinutes: 1440 })] },
      isLoading: false,
      isError: false,
      refetch,
    };

    const { container } = renderPanel({ estimateHours: 8 });

    expect(screen.getByText('flow.time.overEstimate')).toBeInTheDocument();
    const fill = container.querySelector('[role="progressbar"] > div') as HTMLElement;
    expect(fill.style.width).toBe('100%');
  });
});
