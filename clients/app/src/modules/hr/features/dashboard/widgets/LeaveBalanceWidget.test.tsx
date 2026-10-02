import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

import LeaveBalanceWidget from './LeaveBalanceWidget';

describe('LeaveBalanceWidget', () => {
  it('shows the compact empty state and calls onRefresh when data is null', () => {
    const onRefresh = vi.fn();

    render(<LeaveBalanceWidget data={null} isLoading={false} onRefresh={onRefresh} />);

    expect(screen.getByText('hrDashboard.widgets.noLeaveTypes')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'hrCommon.refresh' }));

    expect(onRefresh).toHaveBeenCalledOnce();
  });

  it('shows the compact empty state when data is an empty array', () => {
    const onRefresh = vi.fn();

    render(<LeaveBalanceWidget data={[]} isLoading={false} onRefresh={onRefresh} />);

    expect(screen.getByText('hrDashboard.widgets.noLeaveTypes')).toBeInTheDocument();
  });

  it('renders leave balances instead of the empty state when data is present', () => {
    const onRefresh = vi.fn();

    render(
      <LeaveBalanceWidget
        data={[{ type: 'Vacation', used: 4, entitled: 20 }]}
        isLoading={false}
        onRefresh={onRefresh}
      />
    );

    expect(screen.queryByText('hrDashboard.widgets.noLeaveTypes')).not.toBeInTheDocument();
    expect(screen.getByText('Vacation')).toBeInTheDocument();
  });

  it('shows a loading skeleton instead of the empty state while isLoading is true', () => {
    const onRefresh = vi.fn();

    render(<LeaveBalanceWidget data={null} isLoading={true} onRefresh={onRefresh} />);

    expect(screen.queryByText('hrDashboard.widgets.noLeaveTypes')).not.toBeInTheDocument();
  });
});
