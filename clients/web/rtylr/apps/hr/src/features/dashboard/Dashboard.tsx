import React, { useMemo, useCallback, useState } from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { motion } from 'motion/react';
import type { Variants } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { Link, useNavigate } from '@tanstack/react-router';
import {
  Users,
  Clock,
  Briefcase,
  DollarSign,
  UserPlus,
  CalendarPlus,
  Play,
  ArrowUpRight,
  TrendingUp,
  LogIn,
  LogOut,
  Sparkles,
  AlertTriangle,
  Timer,
  GripVertical,
  Activity,
} from 'lucide-react';
import { useUserStore } from '@voxire/auth';
import { useGetDashboard, useUpdateWidgetLayout } from './useDashboard';
import { useHRDashboard } from './useDashboard';
import { useStaff } from '../staff/useStaff';
import {
  useExpiringDocuments,
  useOvertimeAlerts,
  useHrSettings,
} from '../../shared/hooks/useHrExtras';
import { canViewCompensation } from '../../shared/lib/hrPermissions';
import { ErrorBoundary, EmptyState } from '@voxire/ui';
import { HelmetContainer } from '@voxire/components';
import BirthdayWidget from '../../shared/components/BirthdayWidget';
import AnniversaryWidget from '../../shared/components/AnniversaryWidget';
import { Card, StatCard, useIsMobile, Skeleton } from '../../shared/components/ui';
import HoursThisWeekWidget from './widgets/HoursThisWeekWidget';
import LeaveBalanceWidget from './widgets/LeaveBalanceWidget';
import UpcomingShiftsWidget from './widgets/UpcomingShiftsWidget';
import PendingApprovalsWidget from './widgets/PendingApprovalsWidget';
import TeamAttendanceTodayWidget from './widgets/TeamAttendanceTodayWidget';
import HeadcountWidget from './widgets/HeadcountWidget';
import PayrollCostWidget from './widgets/PayrollCostWidget';
import SetupCompletionWidget from './widgets/SetupCompletionWidget';
import type { DashboardWidget } from '../../shared/schemas/dashboard';
import { safeRatio, toFiniteNumber } from '@voxire/api';
import { formatCurrency } from '../../shared/lib/hrNumberUtils';

const ACCENT = 'var(--hr-accent)';

const DEFAULT_WIDGETS: DashboardWidget[] = [
  {
    id: 1,
    ownerId: 0,
    roleId: 0,
    widgetType: 'headcount',
    positionX: 0,
    positionY: 0,
    width: 4,
    height: 1,
    config: {},
    createdAt: 0,
    updatedAt: 0,
  },
  {
    id: 2,
    ownerId: 0,
    roleId: 0,
    widgetType: 'team_attendance',
    positionX: 4,
    positionY: 0,
    width: 4,
    height: 1,
    config: {},
    createdAt: 0,
    updatedAt: 0,
  },
  {
    id: 3,
    ownerId: 0,
    roleId: 0,
    widgetType: 'hours_this_week',
    positionX: 8,
    positionY: 0,
    width: 4,
    height: 1,
    config: {},
    createdAt: 0,
    updatedAt: 0,
  },
  {
    id: 4,
    ownerId: 0,
    roleId: 0,
    widgetType: 'pending_approvals',
    positionX: 0,
    positionY: 1,
    width: 4,
    height: 1,
    config: {},
    createdAt: 0,
    updatedAt: 0,
  },
  {
    id: 5,
    ownerId: 0,
    roleId: 0,
    widgetType: 'leave_balance',
    positionX: 4,
    positionY: 1,
    width: 4,
    height: 1,
    config: {},
    createdAt: 0,
    updatedAt: 0,
  },
  {
    id: 6,
    ownerId: 0,
    roleId: 0,
    widgetType: 'upcoming_shifts',
    positionX: 8,
    positionY: 1,
    width: 4,
    height: 1,
    config: {},
    createdAt: 0,
    updatedAt: 0,
  },
  {
    id: 7,
    ownerId: 0,
    roleId: 0,
    widgetType: 'payroll_cost',
    positionX: 0,
    positionY: 2,
    width: 6,
    height: 1,
    config: {},
    createdAt: 0,
    updatedAt: 0,
  },
  {
    id: 8,
    ownerId: 0,
    roleId: 0,
    widgetType: 'setup_completion',
    positionX: 6,
    positionY: 2,
    width: 6,
    height: 1,
    config: {},
    createdAt: 0,
    updatedAt: 0,
  },
];

function parseWidgetData<T>(data: Record<string, unknown>, widgetType: string): T | null {
  const entry = data[widgetType];
  if (!entry || typeof entry !== 'object') return null;
  const payload = (entry as { data?: unknown }).data;
  if (!payload) return null;
  return payload as T;
}

function getInitials(name: string): string {
  return name
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);
}

const container: Variants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.07 } },
};
const item: Variants = {
  hidden: { opacity: 0, y: 12 },
  show: { opacity: 1, y: 0, transition: { duration: 0.22, ease: 'easeOut' } },
};

const quickActionDefs = [
  {
    i18nKey: 'dashboard.quickActions.addStaff',
    icon: UserPlus,
    color: ACCENT,
    bg: 'var(--accent-muted)',
    path: '/employees/new',
  },
  {
    i18nKey: 'dashboard.quickActions.scheduleShift',
    icon: CalendarPlus,
    color: 'var(--hr-status-approved)',
    bg: 'color-mix(in srgb, var(--hr-status-approved) 12%, transparent)',
    path: '/shifts/roster',
  },
  {
    i18nKey: 'dashboard.quickActions.runPayroll',
    icon: Play,
    color: 'var(--hr-shift-blue)',
    bg: 'color-mix(in srgb, var(--hr-shift-blue) 12%, transparent)',
    path: '/payroll',
  },
];

const Dashboard: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const isMobile = useIsMobile(900);
  const { data: settings } = useHrSettings();
  const showCompensation = canViewCompensation();

  const {
    data: dashboardResponse,
    isLoading: widgetsLoading,
    refetch: refetchDashboardWidgets,
  } = useGetDashboard();
  const {
    data: legacyDashboard,
    isLoading: statsLoading,
    refetch: refetchLegacyDashboard,
  } = useHRDashboard();
  const { data: staff = [] } = useStaff();
  const { data: expiring = [] } = useExpiringDocuments(30);
  const { data: overtimeAlerts = [] } = useOvertimeAlerts();
  const updateLayout = useUpdateWidgetLayout();

  const loading = widgetsLoading || statsLoading;

  const widgets = useMemo((): DashboardWidget[] => {
    const source =
      dashboardResponse?.widgets && dashboardResponse.widgets.length > 0
        ? dashboardResponse.widgets
        : DEFAULT_WIDGETS;
    // Defense in depth: a stored layout (default or previously saved) must
    // never surface the payroll widget to a role that can't view compensation,
    // even if the layout itself predates this permission check.
    return showCompensation ? source : source.filter((w) => w.widgetType !== 'payroll_cost');
  }, [dashboardResponse, showCompensation]);

  const widgetData = useMemo((): Record<string, unknown> => {
    return (dashboardResponse?.data ?? {}) as Record<string, unknown>;
  }, [dashboardResponse]);

  const [draggedId, setDraggedId] = useState<number | null>(null);

  const handleDragStart = useCallback((e: React.DragEvent, widgetId: number): void => {
    setDraggedId(widgetId);
    e.dataTransfer.effectAllowed = 'move';
  }, []);

  const handleDrop = useCallback(
    (e: React.DragEvent, targetId: number): void => {
      e.preventDefault();
      if (draggedId === null || draggedId === targetId) return;
      const newWidgets = [...widgets];
      const dragIdx = newWidgets.findIndex((w) => w.id === draggedId);
      const targetIdx = newWidgets.findIndex((w) => w.id === targetId);
      if (dragIdx === -1 || targetIdx === -1) return;

      const dragWidget = newWidgets[dragIdx];
      const targetWidget = newWidgets[targetIdx];
      const tmpX = dragWidget.positionX;
      const tmpY = dragWidget.positionY;
      newWidgets[dragIdx] = {
        ...dragWidget,
        positionX: targetWidget.positionX,
        positionY: targetWidget.positionY,
      };
      newWidgets[targetIdx] = { ...targetWidget, positionX: tmpX, positionY: tmpY };

      updateLayout.mutate(newWidgets);
      setDraggedId(null);
    },
    [draggedId, widgets, updateLayout]
  );

  const handleDragOver = useCallback((e: React.DragEvent): void => {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
  }, []);

  const dashData = {
    totalStaff: toFiniteNumber(legacyDashboard?.totalStaff),
    onShiftToday: toFiniteNumber(legacyDashboard?.onShiftToday),
    openPositions: toFiniteNumber(legacyDashboard?.openPositions),
    monthlyPayroll: toFiniteNumber(legacyDashboard?.monthlyPayroll),
  };
  const recentActivity = legacyDashboard?.recentActivity || [];
  const todayAtGlance = legacyDashboard?.todayAtGlance || {
    shiftsScheduled: 0,
    newApplications: 0,
    pendingApprovals: 0,
  };
  const recentHires = staff.slice(-3);
  const avgTenure =
    staff.length > 0 ? (Math.round(staff.length * 0.8) / staff.length).toFixed(1) : '-';

  const stats = [
    {
      label: t('dashboard.stats.totalStaff'),
      value: loading ? '…' : String(dashData.totalStaff),
      change: t('dashboard.stats.thisMonth'),
      trend: 'up' as const,
      icon: Users,
      color: 'var(--hr-accent)',
    },
    {
      label: t('dashboard.stats.onShift'),
      value: loading ? '…' : String(dashData.onShiftToday),
      change:
        dashData.totalStaff > 0
          ? `${(safeRatio(dashData.onShiftToday, dashData.totalStaff) * 100).toFixed(1)}%`
          : '-',
      trend: 'neutral' as const,
      icon: Clock,
      color: 'var(--hr-status-approved)',
    },
    {
      label: t('dashboard.stats.openPositions'),
      value: loading ? '…' : String(dashData.openPositions),
      trend: 'down' as const,
      change: '',
      icon: Briefcase,
      color: 'var(--hr-status-pending)',
    },
    ...(showCompensation
      ? [
          {
            label: t('dashboard.stats.monthlyPayroll'),
            value: loading ? '…' : formatCurrency(dashData.monthlyPayroll, settings?.currency),
            change: '',
            trend: 'up' as const,
            icon: DollarSign,
            color: 'var(--hr-shift-blue)',
          },
        ]
      : []),
  ];

  const renderWidget = (w: DashboardWidget): React.ReactNode => {
    const wData = widgetData;
    switch (w.widgetType) {
      case 'headcount':
        return <HeadcountWidget data={parseWidgetData(wData, 'headcount')} isLoading={loading} />;
      case 'team_attendance':
        return (
          <TeamAttendanceTodayWidget
            data={parseWidgetData(wData, 'team_attendance')}
            isLoading={loading}
          />
        );
      case 'hours_this_week':
        return (
          <HoursThisWeekWidget
            data={parseWidgetData(wData, 'hours_this_week')}
            isLoading={loading}
          />
        );
      case 'pending_approvals':
        return (
          <PendingApprovalsWidget
            data={parseWidgetData(wData, 'pending_approvals')}
            isLoading={loading}
          />
        );
      case 'leave_balance':
        return (
          <LeaveBalanceWidget
            data={parseWidgetData(wData, 'leave_balance')}
            isLoading={loading}
            onRefresh={() => refetchDashboardWidgets()}
          />
        );
      case 'upcoming_shifts':
        return (
          <UpcomingShiftsWidget
            data={parseWidgetData(wData, 'upcoming_shifts')}
            isLoading={loading}
            onRefresh={() => refetchDashboardWidgets()}
          />
        );
      case 'payroll_cost':
        // Defense in depth: even if a widget layout somehow carries this
        // widget for a role that can't view compensation, never render it.
        if (!showCompensation) return null;
        return (
          <PayrollCostWidget data={parseWidgetData(wData, 'payroll_cost')} isLoading={loading} />
        );
      case 'setup_completion':
        return (
          <SetupCompletionWidget
            data={parseWidgetData(wData, 'setup_completion')}
            isLoading={loading}
          />
        );
      default:
        return null;
    }
  };

  return (
    <>
      <HelmetContainer title={pageTitle(t('nav.dashboard'))} />
      <motion.div
        variants={container}
        initial="hidden"
        animate="show"
        style={{ display: 'flex', flexDirection: 'column', gap: 24 }}
      >
        {/* Stat cards */}
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))',
            gap: 16,
          }}
        >
          {stats.map((s) => (
            <motion.div
              key={s.label}
              variants={item}
              style={{
                background: 'var(--hr-card)',
                border: '1px solid var(--hr-border)',
                borderRadius: 12,
                padding: 18,
                display: 'flex',
                flexDirection: 'column',
                gap: 12,
              }}
            >
              <div
                style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}
              >
                <span style={{ fontSize: 12, color: 'var(--hr-text-muted)', fontWeight: 500 }}>
                  {s.label}
                </span>
                <span
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    height: 32,
                    width: 32,
                    borderRadius: 8,
                    background: `${s.color}18`,
                    color: s.color,
                  }}
                >
                  <s.icon size={16} />
                </span>
              </div>
              <div>
                <p
                  style={{
                    fontSize: 26,
                    fontWeight: 700,
                    color: 'var(--hr-text)',
                    margin: 0,
                    lineHeight: 1.1,
                  }}
                >
                  {s.value}
                </p>
                {s.change && (
                  <p
                    style={{
                      fontSize: 11,
                      color:
                        s.trend === 'up'
                          ? 'var(--hr-status-approved)'
                          : s.trend === 'down'
                            ? 'var(--hr-error)'
                            : 'var(--hr-text-muted)',
                      margin: '4px 0 0',
                      display: 'flex',
                      alignItems: 'center',
                      gap: 3,
                    }}
                  >
                    {s.trend === 'up' && <TrendingUp size={11} />}
                    {s.change}
                  </p>
                )}
              </div>
            </motion.div>
          ))}
        </div>

        {/* Draggable widget grid */}
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: isMobile ? '1fr' : 'repeat(12, 1fr)',
            gap: 16,
          }}
        >
          {widgets
            .slice()
            .sort((a, b) => a.positionY - b.positionY || a.positionX - b.positionX)
            .map((w) => (
              <motion.div
                key={w.id}
                variants={item}
                draggable
                onDragStart={(e) => handleDragStart(e as unknown as React.DragEvent, w.id)}
                onDragOver={handleDragOver}
                onDrop={(e) => handleDrop(e as unknown as React.DragEvent, w.id)}
                style={{
                  gridColumn: isMobile ? 'span 1' : `span ${Math.min(w.width, 12)}`,
                  cursor: 'grab',
                  position: 'relative',
                  opacity: draggedId === w.id ? 0.5 : 1,
                }}
              >
                <span
                  style={{
                    position: 'absolute',
                    top: 8,
                    right: 8,
                    color: 'var(--hr-text-dim)',
                    zIndex: 1,
                    opacity: 0.5,
                  }}
                  aria-label={t('dashboard.widgets.dragToReorder')}
                >
                  <GripVertical size={12} />
                </span>
                <ErrorBoundary
                  fallback={
                    <div
                      style={{
                        background: 'var(--hr-card)',
                        border: '1px solid var(--hr-border)',
                        borderRadius: 12,
                        padding: 16,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        height: '100%',
                        color: 'var(--hr-error)',
                        fontSize: 12,
                        fontWeight: 500,
                      }}
                    >
                      {t('dashboard.widgets.errorFallback')}
                    </div>
                  }
                >
                  {renderWidget(w)}
                </ErrorBoundary>
              </motion.div>
            ))}
        </div>

        {/* Activity + Quick actions */}
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: isMobile ? '1fr' : 'minmax(0, 1fr) 320px',
            gap: 16,
          }}
        >
          {/* Recent activity */}
          <motion.div
            variants={item}
            style={{
              background: 'var(--hr-card)',
              border: '1px solid var(--hr-border)',
              borderRadius: 12,
              overflow: 'hidden',
            }}
          >
            <div
              style={{
                padding: '16px 20px',
                borderBottom: '1px solid var(--hr-border)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                gap: 8,
                flexWrap: 'wrap',
              }}
            >
              <h2 style={{ fontSize: 13, fontWeight: 600, color: 'var(--hr-text)', margin: 0 }}>
                {t('dashboard.recentActivity')}
              </h2>
              <span
                style={{
                  fontSize: 11,
                  color: ACCENT,
                  cursor: 'pointer',
                  display: 'flex',
                  alignItems: 'center',
                  gap: 3,
                }}
              >
                {t('dashboard.viewAll')} <ArrowUpRight size={11} />
              </span>
            </div>
            <div>
              {recentActivity.length === 0 && !loading && (
                <EmptyState
                  icon={Activity}
                  title={t('dashboard.noActivity')}
                  action={{
                    label: t('common.refresh'),
                    onClick: () => refetchLegacyDashboard(),
                  }}
                  size="compact"
                />
              )}
              {loading && (
                <div
                  style={{
                    padding: '12px 20px',
                    display: 'flex',
                    flexDirection: 'column',
                    gap: 12,
                  }}
                >
                  {Array.from({ length: 3 }).map((_, i) => (
                    <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                      <Skeleton height={34} width={34} style={{ borderRadius: '50%' }} />
                      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 4 }}>
                        <Skeleton height={13} width="60%" />
                        <Skeleton height={10} width="40%" />
                      </div>
                    </div>
                  ))}
                </div>
              )}
              {recentActivity.map((a, i) => (
                <div
                  key={i}
                  style={{
                    display: 'flex',
                    alignItems: isMobile ? 'flex-start' : 'center',
                    flexDirection: isMobile ? 'column' : 'row',
                    gap: 12,
                    padding: '12px 20px',
                    borderBottom:
                      i < recentActivity.length - 1 ? '1px solid var(--hr-border)' : 'none',
                  }}
                >
                  <span
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      height: 34,
                      width: 34,
                      borderRadius: '50%',
                      background: 'var(--accent-muted)',
                      color: ACCENT,
                      fontSize: 12,
                      fontWeight: 700,
                      flexShrink: 0,
                    }}
                  >
                    {getInitials(a.name)}
                  </span>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <p
                      style={{
                        margin: 0,
                        fontSize: 13,
                        fontWeight: 600,
                        color: 'var(--hr-text)',
                      }}
                    >
                      {a.name}
                    </p>
                    <p style={{ margin: '2px 0 0', fontSize: 11, color: 'var(--hr-text-muted)' }}>
                      {a.role}
                    </p>
                  </div>
                  <div
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 6,
                      flexShrink: 0,
                      flexWrap: 'wrap',
                    }}
                  >
                    <span
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: 4,
                        fontSize: 11,
                        fontWeight: 500,
                        color:
                          a.type === 'in'
                            ? 'var(--hr-status-approved)'
                            : 'var(--hr-status-pending)',
                      }}
                    >
                      {a.type === 'in' ? <LogIn size={11} /> : <LogOut size={11} />}
                      {a.action}
                    </span>
                    <span style={{ fontSize: 11, color: 'var(--hr-text-dim)' }}>{a.time}</span>
                  </div>
                </div>
              ))}
            </div>
          </motion.div>

          {/* Quick actions + today at a glance */}
          <motion.div
            variants={item}
            style={{
              background: 'var(--hr-card)',
              border: '1px solid var(--hr-border)',
              borderRadius: 12,
              overflow: 'hidden',
            }}
          >
            <div style={{ padding: '16px 20px', borderBottom: '1px solid var(--hr-border)' }}>
              <h2 style={{ fontSize: 13, fontWeight: 600, color: 'var(--hr-text)', margin: 0 }}>
                {t('dashboard.quickActionsTitle')}
              </h2>
            </div>
            <div style={{ padding: 12, display: 'flex', flexDirection: 'column', gap: 8 }}>
              {quickActionDefs.map((a) => (
                <Link
                  key={a.i18nKey}
                  to={a.path}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 12,
                    padding: '14px 16px',
                    borderRadius: 10,
                    background: a.bg,
                    textDecoration: 'none',
                    transition: 'opacity 0.15s',
                  }}
                >
                  <a.icon size={18} style={{ color: a.color, flexShrink: 0 }} />
                  <span style={{ fontSize: 13, fontWeight: 600, color: a.color }}>
                    {t(a.i18nKey)}
                  </span>
                </Link>
              ))}
            </div>

            {/* Today at a glance */}
            <div
              style={{
                margin: '0 12px 12px',
                padding: '14px 16px',
                borderRadius: 10,
                background: 'var(--hr-hover)',
                border: '1px solid var(--hr-border)',
              }}
            >
              <p
                style={{
                  margin: '0 0 8px',
                  fontSize: 11,
                  fontWeight: 600,
                  textTransform: 'uppercase',
                  letterSpacing: '0.06em',
                  color: 'var(--hr-text-dim)',
                }}
              >
                {t('dashboard.todayAtGlance')}
              </p>
              {[
                {
                  label: t('dashboard.glance.shiftsScheduled'),
                  value: loading ? '…' : String(todayAtGlance.shiftsScheduled),
                },
                {
                  label: t('dashboard.glance.newApplications'),
                  value: loading ? '…' : String(todayAtGlance.newApplications),
                },
                {
                  label: t('dashboard.glance.pendingApprovals'),
                  value: loading ? '…' : String(todayAtGlance.pendingApprovals),
                },
              ].map((row) => (
                <div
                  key={row.label}
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    marginTop: 6,
                  }}
                >
                  <span style={{ fontSize: 12, color: 'var(--hr-text-muted)' }}>{row.label}</span>
                  <span style={{ fontSize: 13, fontWeight: 600, color: 'var(--hr-text)' }}>
                    {row.value}
                  </span>
                </div>
              ))}
            </div>
          </motion.div>
        </div>

        {/* HR metrics */}
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))',
            gap: 16,
          }}
        >
          <StatCard
            label={t('dashboard.metrics.headcount')}
            value={staff.length}
            sub={t('dashboard.metrics.totalActive')}
            icon={<Users size={14} />}
          />
          <StatCard
            label={t('dashboard.metrics.avgTenure')}
            value={avgTenure !== '-' ? `${avgTenure} ${t('dashboard.metrics.years')}` : '-'}
            sub={t('dashboard.metrics.allStaff')}
            icon={<Clock size={14} />}
            color="var(--hr-status-approved)"
          />
          <StatCard
            label={t('dashboard.metrics.openPositions')}
            value={legacyDashboard?.openPositions ?? 0}
            sub={t('dashboard.metrics.beingFilled')}
            icon={<Briefcase size={14} />}
            color="var(--hr-status-pending)"
          />
        </div>

        {/* Shared widgets row */}
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))',
            gap: 16,
          }}
        >
          <BirthdayWidget />
          <AnniversaryWidget />
          <Card>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 10 }}>
              <span
                style={{
                  padding: 6,
                  borderRadius: 8,
                  background: 'color-mix(in srgb, var(--hr-status-approved) 15%, transparent)',
                  color: 'var(--hr-status-approved)',
                }}
              >
                <Sparkles size={14} />
              </span>
              <h3 style={{ margin: 0, fontSize: 13, fontWeight: 700, color: 'var(--hr-text)' }}>
                {t('dashboard.recentHires')}
              </h3>
            </div>
            {recentHires.length === 0 ? (
              <EmptyState
                icon={UserPlus}
                title={t('dashboard.noRecentHires')}
                action={{
                  label: t('staff.inviteStaff'),
                  onClick: () => navigate({ to: '/employees/new' }),
                }}
                size="compact"
              />
            ) : (
              <ul
                style={{
                  listStyle: 'none',
                  padding: 0,
                  margin: 0,
                  display: 'flex',
                  flexDirection: 'column',
                  gap: 6,
                }}
              >
                {recentHires.map((s) => (
                  <li key={s.id} style={{ fontSize: 12 }}>
                    <Link
                      to="/staff/$id"
                      params={{ id: String(s.id) }}
                      style={{ textDecoration: 'none', color: 'var(--hr-text)' }}
                    >
                      <div style={{ fontWeight: 600 }}>{s.name}</div>
                      <div style={{ fontSize: 10, color: 'var(--hr-text-muted)' }}>{s.role}</div>
                    </Link>
                  </li>
                ))}
              </ul>
            )}
          </Card>
        </div>

        {/* Alerts row */}
        {(expiring.length > 0 || overtimeAlerts.length > 0) && (
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))',
              gap: 16,
            }}
          >
            {expiring.length > 0 && (
              <Card
                style={{
                  borderColor: 'color-mix(in srgb, var(--hr-status-pending) 40%, transparent)',
                  background: 'color-mix(in srgb, var(--hr-status-pending) 6%, transparent)',
                }}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
                  <AlertTriangle size={14} style={{ color: 'var(--hr-status-pending)' }} />
                  <h3 style={{ margin: 0, fontSize: 13, fontWeight: 700, color: 'var(--hr-text)' }}>
                    {t('dashboard.alerts.expiringDocuments')}
                  </h3>
                </div>
                <p style={{ margin: 0, fontSize: 11, color: 'var(--hr-text-muted)' }}>
                  {t('dashboard.alerts.expiringCount', { count: expiring.length })}
                </p>
              </Card>
            )}
            {overtimeAlerts.length > 0 && (
              <Card
                style={{
                  borderColor: 'color-mix(in srgb, var(--hr-error) 40%, transparent)',
                  background: 'color-mix(in srgb, var(--hr-error) 6%, transparent)',
                }}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
                  <Timer size={14} style={{ color: 'var(--hr-error)' }} />
                  <h3 style={{ margin: 0, fontSize: 13, fontWeight: 700, color: 'var(--hr-text)' }}>
                    {t('dashboard.alerts.overtimeAlerts')}
                  </h3>
                </div>
                <p style={{ margin: 0, fontSize: 11, color: 'var(--hr-text-muted)' }}>
                  {t('dashboard.alerts.overtimeCount', { count: overtimeAlerts.length })}
                </p>
              </Card>
            )}
          </div>
        )}
      </motion.div>
    </>
  );
};

export default Dashboard;
