import React from 'react';
import { render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  useAnniversaries: vi.fn(),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('../hooks/useHrExtras', () => ({
  useAnniversaries: mocks.useAnniversaries,
}));

import AnniversaryWidget from './AnniversaryWidget';

const anniversary = (id: number, date: string, years = 3) => ({
  id,
  staffId: id,
  name: `Employee ${id}`,
  date,
  years,
});

describe('AnniversaryWidget day counting', () => {
  beforeEach(() => {
    // Mid-afternoon, so the old timestamp comparison saw a same-day
    // anniversary as already past and rolled it to next year.
    vi.useFakeTimers();
    vi.setSystemTime(new Date(2026, 8, 21, 14, 30));
    mocks.useAnniversaries.mockReset();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('shows Today for an anniversary that falls on the current day', () => {
    mocks.useAnniversaries.mockReturnValue({ data: [anniversary(1, '09-21')] });

    render(<AnniversaryWidget />);

    expect(screen.getByText('Today')).toBeInTheDocument();
    expect(screen.queryByText('in 365d')).not.toBeInTheDocument();
  });

  it('sorts a same-day anniversary ahead of later ones instead of dropping it', () => {
    mocks.useAnniversaries.mockReturnValue({
      data: [
        anniversary(1, '09-22'),
        anniversary(2, '09-21'),
        anniversary(3, '09-23'),
        anniversary(4, '09-24'),
        anniversary(5, '09-25'),
      ],
    });

    render(<AnniversaryWidget />);

    // The list is sliced to four, so a same-day anniversary pushed to the
    // bottom by a 365-day count would disappear entirely.
    expect(screen.getByText('Today')).toBeInTheDocument();
    expect(screen.getByText('Employee 2')).toBeInTheDocument();
    expect(screen.queryByText('Employee 5')).not.toBeInTheDocument();
  });

  it('counts a future anniversary in days', () => {
    mocks.useAnniversaries.mockReturnValue({ data: [anniversary(1, '09-24')] });

    render(<AnniversaryWidget />);

    expect(screen.getByText('in 3d')).toBeInTheDocument();
  });
});
