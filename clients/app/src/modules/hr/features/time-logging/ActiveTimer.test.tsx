import React from 'react';
import { render, screen, act, fireEvent } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

import ActiveTimer from './ActiveTimer';

const TIMER_KEY = 'rtylr:hr:active-timer';

const storedRecord = (description = 'Writing tests'): string =>
  JSON.stringify({ startedAt: Date.now(), description });

describe('ActiveTimer', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  // The record lives in localStorage, so another tab's Stop removes it without
  // telling this one. The interval used to ignore the missing record and leave
  // the running card up with a Stop button that did nothing.
  it('returns to the idle input when the stored record disappears', () => {
    localStorage.setItem(TIMER_KEY, storedRecord());
    render(<ActiveTimer onStop={vi.fn()} />);

    expect(screen.getByText('hrTimeLogging.stopTimer')).toBeInTheDocument();

    localStorage.removeItem(TIMER_KEY);
    act(() => {
      vi.advanceTimersByTime(1000);
    });

    expect(screen.queryByText('hrTimeLogging.stopTimer')).not.toBeInTheDocument();
    expect(screen.getByText('hrTimeLogging.startTimer')).toBeInTheDocument();
  });

  it('reconciles on the storage event without waiting for the next tick', () => {
    localStorage.setItem(TIMER_KEY, storedRecord());
    render(<ActiveTimer onStop={vi.fn()} />);

    localStorage.removeItem(TIMER_KEY);
    act(() => {
      window.dispatchEvent(new StorageEvent('storage', { key: TIMER_KEY }));
    });

    expect(screen.queryByText('hrTimeLogging.stopTimer')).not.toBeInTheDocument();
    expect(screen.getByText('hrTimeLogging.startTimer')).toBeInTheDocument();
  });

  it('clears the running card when Stop finds the record already gone', () => {
    localStorage.setItem(TIMER_KEY, storedRecord());
    render(<ActiveTimer onStop={vi.fn()} />);

    localStorage.removeItem(TIMER_KEY);
    fireEvent.click(screen.getByText('hrTimeLogging.stopTimer'));

    expect(screen.queryByText('hrTimeLogging.stopTimer')).not.toBeInTheDocument();
    expect(screen.getByText('hrTimeLogging.startTimer')).toBeInTheDocument();
  });
});
