import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

import UpcomingShiftsWidget from './UpcomingShiftsWidget';

describe('UpcomingShiftsWidget', () => {
  it('shows the compact empty state and calls onRefresh when data is null', () => {
    const onRefresh = vi.fn();

    render(<UpcomingShiftsWidget data={null} isLoading={false} onRefresh={onRefresh} />);

    expect(screen.getByText('dashboard.widgets.noShifts')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'common.refresh' }));

    expect(onRefresh).toHaveBeenCalledOnce();
  });

  it('shows the compact empty state when data is an empty array', () => {
    const onRefresh = vi.fn();

    render(<UpcomingShiftsWidget data={[]} isLoading={false} onRefresh={onRefresh} />);

    expect(screen.getByText('dashboard.widgets.noShifts')).toBeInTheDocument();
  });

  it('renders upcoming shifts instead of the empty state when data is present', () => {
    const onRefresh = vi.fn();

    render(
      <UpcomingShiftsWidget
        data={[{ id: 1, name: 'Alex Chen', start: '9:00', end: '17:00', role: 'Server' }]}
        isLoading={false}
        onRefresh={onRefresh}
      />
    );

    expect(screen.queryByText('dashboard.widgets.noShifts')).not.toBeInTheDocument();
    expect(screen.getByText('Alex Chen')).toBeInTheDocument();
  });

  it('shows a loading skeleton instead of the empty state while isLoading is true', () => {
    const onRefresh = vi.fn();

    render(<UpcomingShiftsWidget data={null} isLoading={true} onRefresh={onRefresh} />);

    expect(screen.queryByText('dashboard.widgets.noShifts')).not.toBeInTheDocument();
  });
});
