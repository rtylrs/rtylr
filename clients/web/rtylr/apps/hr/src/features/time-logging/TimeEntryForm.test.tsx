import React from 'react';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

const createEntryMutate = vi.hoisted(() => vi.fn());

vi.mock('../../shared/hooks/queries/useTimeEntries', () => ({
  useCreateTimeEntry: () => ({ mutate: createEntryMutate, isPending: false }),
}));

vi.mock('../../shared/hooks/useHrExtras', () => ({
  useMyEmployeeId: () => ({ data: 1, isError: false }),
}));

import TimeEntryForm from './TimeEntryForm';

interface Prefill {
  durationMinutes?: number;
  description?: string;
  source?: 'manual' | 'timer';
}

const form = (open: boolean, prefill: Prefill): React.JSX.Element => (
  <TimeEntryForm open={open} onClose={() => {}} prefill={prefill} />
);

describe('TimeEntryForm prefill', () => {
  beforeEach(() => {
    createEntryMutate.mockReset();
  });

  // The form is mounted while closed, so its state initialisers ran once with
  // the empty prefill. Stopping a timer then opening the modal left the fields
  // empty and the timer's minutes never reached the entry.
  it('seeds duration and description from a prefill that arrives with the open', () => {
    const { rerender } = render(form(false, {}));

    rerender(form(true, { durationMinutes: 90, description: 'Timer work', source: 'timer' }));

    expect(screen.getByLabelText('timeLogging.form.hours')).toHaveValue(1);
    expect(screen.getByLabelText('timeLogging.form.minutes')).toHaveValue(30);
    expect(screen.getByLabelText('timeLogging.form.description')).toHaveValue('Timer work');
  });

  it('resets the fields on each open so the previous entry does not reappear', () => {
    const timed: Prefill = { durationMinutes: 90, description: 'Timer work', source: 'timer' };
    const { rerender } = render(form(true, timed));

    expect(screen.getByLabelText('timeLogging.form.description')).toHaveValue('Timer work');

    rerender(form(false, timed));
    rerender(form(true, {}));

    expect(screen.getByLabelText('timeLogging.form.hours')).toHaveValue(null);
    expect(screen.getByLabelText('timeLogging.form.minutes')).toHaveValue(null);
    expect(screen.getByLabelText('timeLogging.form.description')).toHaveValue('');
  });
});
