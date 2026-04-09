import React from 'react';
import { motion } from 'framer-motion';
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
} from 'lucide-react';
import { useHRDashboard } from '../../hooks/useDashboard';

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
  const { data: dashboard, isLoading: loading } = useHRDashboard();

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
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '16px' }}>
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
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 320px', gap: '16px' }}>

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
          <div style={{ padding: '16px 20px', borderBottom: '1px solid var(--hr-border)', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
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
                  alignItems: 'center',
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
                <div style={{ display: 'flex', alignItems: 'center', gap: '6px', flexShrink: 0 }}>
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
    </motion.div>
  );
};

export default Dashboard;
