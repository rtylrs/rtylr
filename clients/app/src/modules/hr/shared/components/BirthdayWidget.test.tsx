import React from 'react';
import { render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  useBirthdays: vi.fn(),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('../hooks/useHrExtras', () => ({
  useBirthdays: mocks.useBirthdays,
}));

import BirthdayWidget from './BirthdayWidget';

const birthday = (id: number, date: string) => ({
  id,
  staffId: id,
  name: `Employee ${id}`,
  date,
});

describe('BirthdayWidget day counting', () => {
  beforeEach(() => {
    // Mid-afternoon, so the old timestamp comparison saw a same-day birthday as
    // already past and rolled it to next year. AnniversaryWidget documents the
    // same hazard; this widget was never given the matching fix.
    vi.useFakeTimers();
    vi.setSystemTime(new Date(2026, 8, 21, 14, 30));
    mocks.useBirthdays.mockReset();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('shows Today for a birthday that falls on the current day', () => {
    mocks.useBirthdays.mockReturnValue({ data: [birthday(1, '09-21')] });

    render(<BirthdayWidget />);

    expect(screen.getByText('Today')).toBeInTheDocument();
    expect(screen.queryByText('in 365d')).not.toBeInTheDocument();
  });

  it('sorts a same-day birthday ahead of later ones instead of dropping it', () => {
    mocks.useBirthdays.mockReturnValue({
      data: [
        birthday(1, '09-22'),
        birthday(2, '09-21'),
        birthday(3, '09-23'),
        birthday(4, '09-24'),
        birthday(5, '09-25'),
      ],
    });

    render(<BirthdayWidget />);

    // The list is sliced to four, so a same-day birthday pushed to the bottom
    // by a 365-day count would disappear entirely.
    expect(screen.getByText('Today')).toBeInTheDocument();
    expect(screen.getByText('Employee 2')).toBeInTheDocument();
    expect(screen.queryByText('Employee 5')).not.toBeInTheDocument();
  });

  it('counts a future birthday in days', () => {
    mocks.useBirthdays.mockReturnValue({ data: [birthday(1, '09-24')] });

    render(<BirthdayWidget />);

    expect(screen.getByText('in 3d')).toBeInTheDocument();
  });

  it('rolls a birthday that already passed this year to the next one', () => {
    mocks.useBirthdays.mockReturnValue({ data: [birthday(1, '09-20')] });

    render(<BirthdayWidget />);

    expect(screen.queryByText('Today')).not.toBeInTheDocument();
    expect(screen.getByText('in 364d')).toBeInTheDocument();
  });
});
