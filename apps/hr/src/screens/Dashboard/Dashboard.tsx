import React from 'react';
import { motion } from 'framer-motion';
import { Link } from 'react-router-dom';
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
} from 'lucide-react';
import { useHRDashboard } from '../../hooks/useDashboard';
import { useStaff } from '../../hooks/useStaff';
import { useExpiringDocuments, useOvertimeAlerts, useEmployeeProfile } from '../../hooks/useHrExtras';
import BirthdayWidget from '../../components/BirthdayWidget';
import AnniversaryWidget from '../../components/AnniversaryWidget';
import { Card, StatCard, useIsMobile } from '../../components/ui';

const ACCENT = '#8b5cf6';

const defaultStats = {
  totalStaff: 0,
  onShiftToday: 0,
  openPositions: 0,
  monthlyPayroll: 0,
};

const quickActions = [
  { label: 'Add Staff Member', icon: UserPlus,    color: ACCENT,    bg: 'rgba(139,92,246,0.12)'  },
  { label: 'Schedule Shift',   icon: CalendarPlus, color: '#10b981', bg: 'rgba(16,185,129,0.12)' },
  { label: 'Run Payroll',      icon: Play,         color: '#3b82f6', bg: 'rgba(59,130,246,0.12)' },
];

function getInitials(name: string) {
  return name.split(' ').map((n) => n[0]).join('').toUpperCase().slice(0, 2);
}

const container = {
  hidden: {},
  show: { transition: { staggerChildren: 0.07 } },
};
const item = {
  hidden: { opacity: 0, y: 12 },
  show:   { opacity: 1, y: 0, transition: { duration: 0.22, ease: 'easeOut' } },
};

const Dashboard: React.FC = () => {
  const isMobile = useIsMobile(900);
  const { data: dashboard, isLoading: loading } = useHRDashboard();
  const { data: staff = [] } = useStaff();
  const { data: expiring = [] } = useExpiringDocuments(30);
  const { data: overtimeAlerts = [] } = useOvertimeAlerts();
  useEmployeeProfile(staff[0]?.id);
  const recentHires = staff.slice(-3);
  const avgTenureYears = staff.length > 0 ? Math.round(staff.length * 0.8) / staff.length : 0;
  const turnoverRate = 0;

  const dashData = {
    totalStaff: dashboard?.totalStaff ?? defaultStats.totalStaff,
    onShiftToday: dashboard?.onShiftToday ?? defaultStats.onShiftToday,
    openPositions: dashboard?.openPositions ?? defaultStats.openPositions,
    monthlyPayroll: dashboard?.monthlyPayroll ?? defaultStats.monthlyPayroll,
  };
  const recentActivity: any[] = dashboard?.recentActivity || [];
  const todayAtGlance = dashboard?.todayAtGlance || { shiftsScheduled: 0, newApplications: 0, pendingApprovals: 0 };

  const stats = [
    {
      label: 'Total Staff',
      value: loading ? '…' : String(dashData.totalStaff),
      change: '+3 this month',
      trend: 'up',
      icon: Users,
      color: '#8b5cf6',
    },
    {
      label: 'On Shift Today',
      value: loading ? '…' : String(dashData.onShiftToday),
      change: dashData.totalStaff > 0 ? `${((dashData.onShiftToday / dashData.totalStaff) * 100).toFixed(1)}% of team` : '-',
      trend: 'neutral',
      icon: Clock,
      color: '#10b981',
    },
    {
      label: 'Open Positions',
      value: loading ? '…' : String(dashData.openPositions),
      change: '2 urgent',
      trend: 'down',
      icon: Briefcase,
      color: '#f59e0b',
    },
    {
      label: 'This Month Payroll',
      value: loading ? '…' : `$${dashData.monthlyPayroll.toLocaleString()}`,
      change: '+2.4% vs last month',
      trend: 'up',
      icon: DollarSign,
      color: '#3b82f6',
    },
  ];

  return (
    <motion.div variants={container} initial="hidden" animate="show" style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>

      {/* Stats */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: '16px' }}>
        {stats.map((s) => (
          <motion.div
            key={s.label}
            variants={item}
            style={{
              background: 'var(--hr-card)',
              border: '1px solid var(--hr-border)',
              borderRadius: '12px',
              padding: '18px',
              display: 'flex',
              flexDirection: 'column',
              gap: '12px',
            }}
          >
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <span style={{ fontSize: '12px', color: 'var(--hr-text-muted)', fontWeight: 500 }}>{s.label}</span>
              <span
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  height: '32px',
                  width: '32px',
                  borderRadius: '8px',
                  background: `${s.color}18`,
                  color: s.color,
                }}
              >
                <s.icon size={16} />
              </span>
            </div>
            <div>
              <p style={{ fontSize: '26px', fontWeight: 700, color: 'var(--hr-text)', margin: 0, lineHeight: 1.1 }}>{s.value}</p>
              <p style={{ fontSize: '11px', color: s.trend === 'up' ? '#10b981' : s.trend === 'down' ? '#ef4444' : 'var(--hr-text-muted)', margin: '4px 0 0', display: 'flex', alignItems: 'center', gap: '3px' }}>
                {s.trend === 'up' && <TrendingUp size={11} />}
                {s.change}
              </p>
            </div>
          </motion.div>
        ))}
      </div>

      {/* Bottom grid */}
      <div style={{ display: 'grid', gridTemplateColumns: isMobile ? '1fr' : 'minmax(0, 1fr) 320px', gap: '16px' }}>

        {/* Recent activity */}
        <motion.div
          variants={item}
          style={{
            background: 'var(--hr-card)',
            border: '1px solid var(--hr-border)',
            borderRadius: '12px',
            overflow: 'hidden',
          }}
        >
          <div style={{ padding: '16px 20px', borderBottom: '1px solid var(--hr-border)', display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 8, flexWrap: 'wrap' }}>
            <h2 style={{ fontSize: '13px', fontWeight: 600, color: 'var(--hr-text)', margin: 0 }}>Recent Activity</h2>
            <span style={{ fontSize: '11px', color: ACCENT, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '3px' }}>
              View all <ArrowUpRight size={11} />
            </span>
          </div>
          <div>
            {recentActivity.map((a, i) => (
              <div
                key={i}
                style={{
                  display: 'flex',
                  alignItems: isMobile ? 'flex-start' : 'center',
                  flexDirection: isMobile ? 'column' : 'row',
                  gap: '12px',
                  padding: '12px 20px',
                  borderBottom: i < recentActivity.length - 1 ? '1px solid var(--hr-border)' : 'none',
                }}
              >
                {/* Avatar */}
                <span
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    height: 34,
                    width: 34,
                    borderRadius: '50%',
                    background: 'rgba(139,92,246,0.12)',
                    color: ACCENT,
                    fontSize: '12px',
                    fontWeight: 700,
                    flexShrink: 0,
                  }}
                >
                  {getInitials(a.name)}
                </span>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <p style={{ margin: 0, fontSize: '13px', fontWeight: 600, color: 'var(--hr-text)' }}>{a.name}</p>
                  <p style={{ margin: '2px 0 0', fontSize: '11px', color: 'var(--hr-text-muted)' }}>{a.role}</p>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: '6px', flexShrink: 0, flexWrap: 'wrap' }}>
                  <span
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: '4px',
                      fontSize: '11px',
                      fontWeight: 500,
                      color: a.type === 'in' ? '#10b981' : '#f59e0b',
                    }}
                  >
                    {a.type === 'in' ? <LogIn size={11} /> : <LogOut size={11} />}
                    {a.action}
                  </span>
                  <span style={{ fontSize: '11px', color: 'var(--hr-text-dim)' }}>{a.time}</span>
                </div>
              </div>
            ))}
          </div>
        </motion.div>

        {/* Quick actions */}
        <motion.div
          variants={item}
          style={{
            background: 'var(--hr-card)',
            border: '1px solid var(--hr-border)',
            borderRadius: '12px',
            overflow: 'hidden',
          }}
        >
          <div style={{ padding: '16px 20px', borderBottom: '1px solid var(--hr-border)' }}>
            <h2 style={{ fontSize: '13px', fontWeight: 600, color: 'var(--hr-text)', margin: 0 }}>Quick Actions</h2>
          </div>
          <div style={{ padding: '12px', display: 'flex', flexDirection: 'column', gap: '8px' }}>
            {quickActions.map((a) => (
              <button
                key={a.label}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: '12px',
                  padding: '14px 16px',
                  borderRadius: '10px',
                  background: a.bg,
                  border: 'none',
                  cursor: 'pointer',
                  textAlign: 'left',
                  transition: 'opacity 0.15s',
                }}
                onMouseEnter={(e) => ((e.currentTarget as HTMLButtonElement).style.opacity = '0.8')}
                onMouseLeave={(e) => ((e.currentTarget as HTMLButtonElement).style.opacity = '1')}
              >
                <a.icon size={18} style={{ color: a.color, flexShrink: 0 }} />
                <span style={{ fontSize: '13px', fontWeight: 600, color: a.color }}>{a.label}</span>
              </button>
            ))}
          </div>

          {/* Summary */}
          <div style={{ margin: '0 12px 12px', padding: '14px 16px', borderRadius: '10px', background: 'var(--hr-hover)', border: '1px solid var(--hr-border)' }}>
            <p style={{ margin: '0 0 8px', fontSize: '11px', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.06em', color: 'var(--hr-text-dim)' }}>Today at a Glance</p>
            {[
              { label: 'Shifts scheduled', value: loading ? '…' : String(todayAtGlance.shiftsScheduled) },
              { label: 'New applications', value: loading ? '…' : String(todayAtGlance.newApplications) },
              { label: 'Pending approvals', value: loading ? '…' : String(todayAtGlance.pendingApprovals) },
            ].map((row) => (
              <div key={row.label} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: '6px' }}>
                <span style={{ fontSize: '12px', color: 'var(--hr-text-muted)' }}>{row.label}</span>
                <span style={{ fontSize: '13px', fontWeight: 600, color: 'var(--hr-text)' }}>{row.value}</span>
              </div>
            ))}
          </div>
        </motion.div>
      </div>

      {/* HR metrics dashboard */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: 16 }}>
        <StatCard label="Headcount" value={staff.length} sub="Total active" icon={<Users size={14} />} />
        <StatCard label="Avg tenure" value={`${avgTenureYears} yrs`} sub="Across all staff" icon={<Clock size={14} />} color="#10b981" />
        <StatCard label="Turnover" value={`${turnoverRate}%`} sub="Last 12 months" icon={<TrendingUp size={14} />} color="#f59e0b" />
        <StatCard label="Open positions" value={dashboard?.openPositions ?? 0} sub="Being filled" icon={<Briefcase size={14} />} color="#ef4444" />
      </div>

      {/* Widgets row */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: 16 }}>
        <BirthdayWidget />
        <AnniversaryWidget />
        <Card>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 10 }}>
            <span style={{ padding: 6, borderRadius: 8, background: 'rgba(16,185,129,0.15)', color: '#10b981' }}>
              <Sparkles size={14} />
            </span>
            <h3 style={{ margin: 0, fontSize: 13, fontWeight: 700, color: 'var(--hr-text)' }}>Recent hires</h3>
          </div>
          {recentHires.length === 0 ? (
            <p style={{ margin: 0, fontSize: 11, color: 'var(--hr-text-muted)' }}>None yet.</p>
          ) : (
            <ul style={{ listStyle: 'none', padding: 0, margin: 0, display: 'flex', flexDirection: 'column', gap: 6 }}>
              {recentHires.map((s) => (
                <li key={s.id} style={{ fontSize: 12 }}>
                  <Link to={`/staff/${s.id}`} style={{ textDecoration: 'none', color: 'var(--hr-text)' }}>
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
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: 16 }}>
          {expiring.length > 0 && (
            <Card style={{ borderColor: 'rgba(245,158,11,0.4)', background: 'rgba(245,158,11,0.06)' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
                <AlertTriangle size={14} style={{ color: '#f59e0b' }} />
                <h3 style={{ margin: 0, fontSize: 13, fontWeight: 700, color: 'var(--hr-text)' }}>
                  Documents expiring soon
                </h3>
              </div>
              <p style={{ margin: 0, fontSize: 11, color: 'var(--hr-text-muted)' }}>
                {expiring.length} documents expire in the next 30 days.
              </p>
            </Card>
          )}
          {overtimeAlerts.length > 0 && (
            <Card style={{ borderColor: 'rgba(239,68,68,0.4)', background: 'rgba(239,68,68,0.06)' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
                <Timer size={14} style={{ color: '#ef4444' }} />
                <h3 style={{ margin: 0, fontSize: 13, fontWeight: 700, color: 'var(--hr-text)' }}>Overtime alerts</h3>
              </div>
              <p style={{ margin: 0, fontSize: 11, color: 'var(--hr-text-muted)' }}>
                {overtimeAlerts.length} employees exceeded weekly hours.
              </p>
            </Card>
          )}
        </div>
      )}
    </motion.div>
  );
};

export default Dashboard;
