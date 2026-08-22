import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { DashboardWidget } from '../../shared/schemas/dashboard';

const mocks = vi.hoisted(() => ({
  useUserStore: vi.fn(),
  useGetDashboard: vi.fn(),
  useHRDashboard: vi.fn(),
  useUpdateWidgetLayout: vi.fn(),
  useStaff: vi.fn(),
  useExpiringDocuments: vi.fn(),
  useOvertimeAlerts: vi.fn(),
  useHrSettings: vi.fn(),
  useHRConfigStore: vi.fn(),
  navigate: vi.fn(),
  refetchDashboardWidgets: vi.fn(),
  refetchLegacyDashboard: vi.fn(),
}));

vi.mock('@voxire/auth', () => ({
  useUserStore: mocks.useUserStore,
}));

vi.mock('./useDashboard', () => ({
  useGetDashboard: mocks.useGetDashboard,
  useHRDashboard: mocks.useHRDashboard,
  useUpdateWidgetLayout: mocks.useUpdateWidgetLayout,
}));

vi.mock('../staff/useStaff', () => ({
  useStaff: mocks.useStaff,
}));

vi.mock('../../shared/hooks/useHrExtras', () => ({
  useExpiringDocuments: mocks.useExpiringDocuments,
  useOvertimeAlerts: mocks.useOvertimeAlerts,
  useHrSettings: mocks.useHrSettings,
}));

vi.mock('../../shared/stores/hrConfigStore', () => ({
  useHRConfigStore: mocks.useHRConfigStore,
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('@voxire/ui', () => ({
  ErrorBoundary: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  EmptyState: ({
    title,
    action,
  }: {
    title: string;
    action: { label: string; onClick: () => void };
  }) => (
    <div>
      <span>{title}</span>
      <button type="button" onClick={action.onClick}>
        {action.label}
      </button>
    </div>
  ),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@tanstack/react-router', () => ({
  Link: ({ children, to }: { children: React.ReactNode; to: string }) => (
    <a href={to}>{children}</a>
  ),
  useNavigate: () => mocks.navigate,
}));

vi.mock('motion/react', () => ({
  motion: {
    div: ({
      children,
      variants: _variants,
      initial: _initial,
      animate: _animate,
      ...rest
    }: React.HTMLAttributes<HTMLDivElement> & Record<string, unknown>) => (
      <div {...rest}>{children as React.ReactNode}</div>
    ),
  },
}));

vi.mock('../../shared/components/BirthdayWidget', () => ({ default: () => null }));
vi.mock('../../shared/components/AnniversaryWidget', () => ({ default: () => null }));
vi.mock('./widgets/HoursThisWeekWidget', () => ({
  default: () => <div data-testid="widget-hours_this_week" />,
}));
vi.mock('./widgets/LeaveBalanceWidget', () => ({
  default: () => <div data-testid="widget-leave_balance" />,
}));
vi.mock('./widgets/UpcomingShiftsWidget', () => ({
  default: () => <div data-testid="widget-upcoming_shifts" />,
}));
vi.mock('./widgets/PendingApprovalsWidget', () => ({
  default: () => <div data-testid="widget-pending_approvals" />,
}));
vi.mock('./widgets/TeamAttendanceTodayWidget', () => ({
  default: () => <div data-testid="widget-team_attendance" />,
}));
vi.mock('./widgets/HeadcountWidget', () => ({
  default: () => <div data-testid="widget-headcount" />,
}));
vi.mock('./widgets/PayrollCostWidget', () => ({
  default: () => <div data-testid="widget-payroll_cost" />,
}));
vi.mock('./widgets/SetupCompletionWidget', () => ({
  default: () => <div data-testid="widget-setup_completion" />,
}));

import Dashboard from './Dashboard';

function role(r: string): void {
  mocks.useUserStore.mockImplementation(
    (selector: (state: { user: { role: string } }) => unknown) => selector({ user: { role: r } })
  );
}

describe('Dashboard compensation gating', () => {
  beforeEach(() => {
    mocks.useUserStore.mockReset();
    mocks.useGetDashboard.mockReset();
    mocks.useHRDashboard.mockReset();
    mocks.useUpdateWidgetLayout.mockReset();
    mocks.useStaff.mockReset();
    mocks.useExpiringDocuments.mockReset();
    mocks.useOvertimeAlerts.mockReset();
    mocks.useHrSettings.mockReset();
    mocks.useHRConfigStore.mockReset();
    mocks.navigate.mockReset();
    mocks.refetchDashboardWidgets.mockReset();
    mocks.refetchLegacyDashboard.mockReset();

    mocks.useGetDashboard.mockReturnValue({
      data: undefined,
      isLoading: false,
      refetch: mocks.refetchDashboardWidgets,
    });
    mocks.useHRDashboard.mockReturnValue({
      data: {
        totalStaff: 10,
        onShiftToday: 4,
        openPositions: 2,
        monthlyPayroll: 42000,
        recentActivity: [],
        todayAtGlance: { shiftsScheduled: 0, newApplications: 0, pendingApprovals: 0 },
      },
      isLoading: false,
      refetch: mocks.refetchLegacyDashboard,
    });
    mocks.useUpdateWidgetLayout.mockReturnValue({ mutate: vi.fn() });
    mocks.useStaff.mockReturnValue({ data: [] });
    mocks.useExpiringDocuments.mockReturnValue({ data: [] });
    mocks.useOvertimeAlerts.mockReturnValue({ data: [] });
    mocks.useHrSettings.mockReturnValue({ data: { currency: 'USD' } });
    mocks.useHRConfigStore.mockImplementation((selector: (state: { config: unknown }) => unknown) =>
      selector({ config: {} })
    );
  });

  it('hides the payroll stat and payroll widget for a staff-role user (default layout)', () => {
    role('staff');

    render(<Dashboard />);

    expect(screen.queryByText('dashboard.stats.monthlyPayroll')).not.toBeInTheDocument();
    expect(screen.queryByTestId('widget-payroll_cost')).not.toBeInTheDocument();
    // Other default widgets are unaffected by the compensation gate.
    expect(screen.getByTestId('widget-headcount')).toBeInTheDocument();
  });

  it('shows the payroll stat and payroll widget for an admin-role user (default layout)', () => {
    role('admin');

    render(<Dashboard />);

    expect(screen.getByText('dashboard.stats.monthlyPayroll')).toBeInTheDocument();
    expect(screen.getByTestId('widget-payroll_cost')).toBeInTheDocument();
  });

  it('shows the payroll stat and payroll widget for an owner-role user (default layout)', () => {
    role('owner');

    render(<Dashboard />);

    expect(screen.getByText('dashboard.stats.monthlyPayroll')).toBeInTheDocument();
    expect(screen.getByTestId('widget-payroll_cost')).toBeInTheDocument();
  });

  it('never renders the payroll widget for a staff-role user even if a stored layout includes it', () => {
    // Simulates a layout persisted before this permission check existed, or a
    // layout shared across roles - defense in depth against a stale/foreign
    // widget list still carrying the payroll_cost widget.
    const storedWidgets: DashboardWidget[] = [
      {
        id: 99,
        ownerId: 0,
        roleId: 0,
        widgetType: 'payroll_cost',
        positionX: 0,
        positionY: 0,
        width: 6,
        height: 1,
        config: {},
        createdAt: 0,
        updatedAt: 0,
      },
      {
        id: 100,
        ownerId: 0,
        roleId: 0,
        widgetType: 'headcount',
        positionX: 6,
        positionY: 0,
        width: 6,
        height: 1,
        config: {},
        createdAt: 0,
        updatedAt: 0,
      },
    ];
    mocks.useGetDashboard.mockReturnValue({
      data: { widgets: storedWidgets, data: {} },
      isLoading: false,
    });
    role('staff');

    render(<Dashboard />);

    expect(screen.queryByTestId('widget-payroll_cost')).not.toBeInTheDocument();
    expect(screen.getByTestId('widget-headcount')).toBeInTheDocument();
  });

  it('shows the recent activity empty state and refetches the legacy dashboard on CTA click', () => {
    role('owner');

    render(<Dashboard />);

    expect(screen.getByText('dashboard.noActivity')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'common.refresh' }));

    expect(mocks.refetchLegacyDashboard).toHaveBeenCalledOnce();
  });

  it('shows the recent hires empty state and navigates to invite staff on CTA click', () => {
    role('owner');
    mocks.useStaff.mockReturnValue({ data: [] });

    render(<Dashboard />);

    expect(screen.getByText('dashboard.noRecentHires')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'staff.inviteStaff' }));

    expect(mocks.navigate).toHaveBeenCalledWith({ to: '/employees/new' });
  });

  it('does not show the recent hires empty state once staff exist', () => {
    role('owner');
    mocks.useStaff.mockReturnValue({
      data: [{ id: 1, name: 'Alex Chen', role: 'Server', department: 'FOH' }],
    });

    render(<Dashboard />);

    expect(screen.queryByText('dashboard.noRecentHires')).not.toBeInTheDocument();
  });
});
