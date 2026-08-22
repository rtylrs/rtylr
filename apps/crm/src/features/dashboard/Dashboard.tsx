import React from 'react';
import { HelmetContainer } from '@voxire/components';
import { useNavigate } from '@tanstack/react-router';
import { motion } from 'motion/react';
import Skeleton from 'react-loading-skeleton';
import 'react-loading-skeleton/dist/skeleton.css';
import {
  Users,
  UserPlus,
  Gift,
  Megaphone,
  TrendingUp,
  TrendingDown,
  AlertTriangle,
  DollarSign,
  PieChart,
  Activity,
  ArrowRight,
} from 'lucide-react';
import { useCrmDashboard } from '../../shared/hooks/queries/useCrmDashboard';
import { useListAllCustomers } from '../../shared/hooks/queries/useCustomers';
import { useListSegments } from '../../shared/hooks/queries/useSegments';
import type { DashboardStats } from '../../shared/schemas/dashboard';
import { CustomerAvatar } from '../../shared/components/crm';
import { activateOnKeyboard } from '../../shared/lib/keyboard';
import { parseCrmDate } from '../../shared/lib/dateUtils';
import { formatCustomerDate } from '../../shared/lib/customerUtils';

const ACCENT = '#3b82f6';

interface StatItem {
  label: string;
  value: string;
  change: string | undefined;
  up: boolean | undefined;
  icon: React.ElementType;
  color: string;
}

const tierColor: Record<string, string> = {
  Platinum: '#e879f9',
  Gold: '#f59e0b',
  Silver: '#94a3b8',
  Bronze: '#b45309',
};

const DAY_MS = 24 * 60 * 60 * 1000;

export function isAtRiskCustomer(lastVisit: string, thresholdDays = 30): boolean {
  const value = lastVisit.trim();
  if (!value) return true;
  const timestamp = parseCrmDate(value).getTime();
  if (Number.isFinite(timestamp)) {
    return Date.now() - timestamp >= Math.max(0, thresholdDays) * DAY_MS;
  }
  const relative = value.match(/(\d+)\s+(minute|hour|day|week|month|year)s?\s+ago/i);
  if (relative) {
    const units: Record<string, number> = {
      minute: 1 / (24 * 60),
      hour: 1 / 24,
      day: 1,
      week: 7,
      month: 30,
      year: 365,
    };
    return Number(relative[1]) * (units[relative[2].toLowerCase()] ?? 0) >= thresholdDays;
  }
  if (/(?:today|yesterday|minute|minutes|hour|hours)\b/i.test(value)) return false;
  if (/(?:never|no visits?|unknown)\b/i.test(value)) return true;
  return /(?:month|months|week|weeks|year|years)\b/i.test(value);
}

function StatCard({ stat, index }: { stat: StatItem; index: number }): React.JSX.Element {
  const Icon = stat.icon;
  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: index * 0.06, duration: 0.22, ease: 'easeOut' }}
      style={{
        background: 'var(--hub-card)',
        border: '1px solid var(--hub-border)',
        borderRadius: '12px',
        padding: '20px',
        display: 'flex',
        flexDirection: 'column',
        gap: '12px',
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <span style={{ fontSize: '12px', fontWeight: 500, color: 'var(--hub-text-muted)' }}>
          {stat.label}
        </span>
        <div
          style={{
            width: '32px',
            height: '32px',
            borderRadius: '8px',
            background: `${stat.color}18`,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          <Icon size={15} style={{ color: stat.color }} />
        </div>
      </div>
      <div style={{ display: 'flex', alignItems: 'flex-end', gap: '8px' }}>
        <span
          style={{ fontSize: '26px', fontWeight: 700, color: 'var(--hub-text)', lineHeight: 1 }}
        >
          {stat.value}
        </span>
        {stat.change ? (
          <span
            style={{
              fontSize: '12px',
              fontWeight: 500,
              color: stat.up ? '#10b981' : '#ef4444',
              display: 'flex',
              alignItems: 'center',
              gap: '2px',
              paddingBottom: '2px',
            }}
          >
            {stat.up ? <TrendingUp size={12} /> : <TrendingDown size={12} />}
            {stat.change}
          </span>
        ) : (
          <span style={{ fontSize: '11px', color: 'var(--hub-text-dim)', paddingBottom: '2px' }}>
            No trend data
          </span>
        )}
      </div>
      <div
        style={{
          height: '3px',
          borderRadius: '9999px',
          background: 'var(--hub-border)',
          overflow: 'hidden',
        }}
      >
        <div
          style={{
            height: '100%',
            width: stat.up ? '62%' : '38%',
            borderRadius: '9999px',
            background: stat.color,
          }}
        />
      </div>
    </motion.div>
  );
}

const Dashboard: React.FC = () => {
  const navigate = useNavigate();
  const {
    data: dashboardData,
    isLoading: dashLoading,
    isError: dashboardError,
  } = useCrmDashboard();
  const {
    data: fetchedCustomers,
    isLoading: customersLoading,
    isError: customersError,
  } = useListAllCustomers();
  const {
    data: fetchedSegments,
    isLoading: segmentsLoading,
    isError: segmentsError,
  } = useListSegments();
  const customersPool = fetchedCustomers ?? [];
  const segmentsPool = fetchedSegments ?? [];

  const ltv =
    customersLoading || customersError
      ? null
      : customersPool.length > 0
        ? Math.round(customersPool.reduce((s, c) => s + c.totalSpent, 0) / customersPool.length)
        : 0;
  const atRiskCustomers = customersPool.filter((c) => isAtRiskCustomer(c.lastVisit));
  const atRisk = atRiskCustomers.slice(0, 4);
  const churnRate =
    customersLoading || customersError
      ? null
      : customersPool.length > 0
        ? Math.round((atRiskCustomers.length / customersPool.length) * 100)
        : 0;
  const topSegment = segmentsError
    ? null
    : segmentsPool.length > 0
      ? segmentsPool.reduce((a, b) => (a.customerCount > b.customerCount ? a : b), segmentsPool[0])
      : null;

  const stats = dashboardData
    ? [
        {
          label: 'Total Customers',
          value: (dashboardData as DashboardStats).totalCustomers.toLocaleString(),
          change: dashboardData.totalCustomersChange,
          up: dashboardData.totalCustomersChange?.startsWith('+'),
          icon: Users,
          color: '#3b82f6',
        },
        {
          label: 'New This Month',
          value: dashboardData.newThisMonth.toLocaleString(),
          change: dashboardData.newThisMonthChange,
          up: dashboardData.newThisMonthChange?.startsWith('+'),
          icon: UserPlus,
          color: '#10b981',
        },
        {
          label: 'Active Loyalty Members',
          value: dashboardData.activeLoyaltyMembers.toLocaleString(),
          change: dashboardData.activeLoyaltyMembersChange,
          up: dashboardData.activeLoyaltyMembersChange?.startsWith('+'),
          icon: Gift,
          color: '#f59e0b',
        },
        {
          label: 'Campaign Reach',
          value:
            dashboardData.campaignReach === undefined
              ? 'Unavailable'
              : dashboardData.campaignReach.toLocaleString(),
          change: dashboardData.campaignReachChange,
          up: dashboardData.campaignReachChange?.startsWith('+'),
          icon: Megaphone,
          color: '#8b5cf6',
        },
      ]
    : [];
  const topCustomers = dashboardData?.topCustomers ?? [];
  const campaigns = dashboardData?.recentCampaigns ?? [];

  return (
    <>
      <HelmetContainer title="Dashboard · rtylr" />
      <div style={{ display: 'flex', flexDirection: 'column', gap: '24px', maxWidth: '1280px' }}>
        {/* Stats grid */}
        <div className="crm-stats-grid" style={{ display: 'grid', gap: '16px' }}>
          {dashLoading ? (
            [0, 1, 2, 3].map((i) => (
              <div
                key={i}
                style={{
                  background: 'var(--hub-card)',
                  border: '1px solid var(--hub-border)',
                  borderRadius: '12px',
                  padding: '16px',
                }}
              >
                <Skeleton height={12} width="60%" style={{ marginBottom: 12 }} />
                <Skeleton height={28} width="50%" style={{ marginBottom: 6 }} />
                <Skeleton height={10} width="40%" />
              </div>
            ))
          ) : dashboardError ? (
            <div
              role="alert"
              style={{
                gridColumn: '1 / -1',
                padding: '24px',
                borderRadius: '12px',
                border: '1px solid rgba(239,68,68,0.25)',
                background: 'rgba(239,68,68,0.08)',
                color: 'var(--hub-text-muted)',
                fontSize: '13px',
              }}
            >
              Unable to load dashboard metrics. Refresh and try again.
            </div>
          ) : (
            stats.map((stat, i) => <StatCard key={stat.label} stat={stat} index={i} />)
          )}
        </div>

        {/* Extended KPI widgets */}
        <div className="crm-stats-grid" style={{ display: 'grid', gap: '16px' }}>
          <motion.div
            initial={{ opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.18 }}
            style={{
              background: 'var(--hub-card)',
              border: '1px solid var(--hub-border)',
              borderRadius: '12px',
              padding: '16px',
            }}
          >
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                marginBottom: '8px',
              }}
            >
              <span
                style={{
                  fontSize: '11px',
                  color: 'var(--hub-text-dim)',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                  fontWeight: 700,
                }}
              >
                Avg LTV
              </span>
              <div
                style={{
                  width: '28px',
                  height: '28px',
                  borderRadius: '8px',
                  background: 'rgba(16,185,129,0.12)',
                  color: '#10b981',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                }}
              >
                <DollarSign size={13} />
              </div>
            </div>
            <div style={{ fontSize: '24px', fontWeight: 700, color: 'var(--hub-text)' }}>
              {ltv === null ? 'Unavailable' : `$${ltv.toLocaleString()}`}
            </div>
            <div style={{ fontSize: '11px', color: 'var(--hub-text-muted)', marginTop: '2px' }}>
              per customer
            </div>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.22 }}
            onClick={() => navigate({ to: '/customers/new' })}
            role="button"
            tabIndex={0}
            onKeyDown={(event) =>
              activateOnKeyboard(event, () => navigate({ to: '/customers/new' }))
            }
            style={{
              background: 'var(--hub-card)',
              border: '1px solid var(--hub-border)',
              borderRadius: '12px',
              padding: '16px',
              cursor: 'pointer',
            }}
          >
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                marginBottom: '8px',
              }}
            >
              <span
                style={{
                  fontSize: '11px',
                  color: 'var(--hub-text-dim)',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                  fontWeight: 700,
                }}
              >
                New This Month
              </span>
              <div
                style={{
                  width: '28px',
                  height: '28px',
                  borderRadius: '8px',
                  background: 'rgba(59,130,246,0.12)',
                  color: '#3b82f6',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                }}
              >
                <UserPlus size={13} />
              </div>
            </div>
            <div style={{ fontSize: '24px', fontWeight: 700, color: 'var(--hub-text)' }}>
              {dashLoading ? (
                <Skeleton width={60} />
              ) : dashboardError ? (
                'Unavailable'
              ) : (
                (dashboardData?.newThisMonth?.toLocaleString() ?? 'Unavailable')
              )}
            </div>
            <div style={{ fontSize: '11px', color: '#10b981', marginTop: '2px' }}>
              {dashboardError
                ? 'Dashboard data unavailable'
                : (dashboardData?.newThisMonthChange ?? 'No trend data')}
            </div>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.26 }}
            style={{
              background: 'var(--hub-card)',
              border: '1px solid var(--hub-border)',
              borderRadius: '12px',
              padding: '16px',
            }}
          >
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                marginBottom: '8px',
              }}
            >
              <span
                style={{
                  fontSize: '11px',
                  color: 'var(--hub-text-dim)',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                  fontWeight: 700,
                }}
              >
                Churn Rate
              </span>
              <div
                style={{
                  width: '28px',
                  height: '28px',
                  borderRadius: '8px',
                  background: 'rgba(239,68,68,0.12)',
                  color: '#ef4444',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                }}
              >
                <Activity size={13} />
              </div>
            </div>
            <div style={{ fontSize: '24px', fontWeight: 700, color: 'var(--hub-text)' }}>
              {churnRate === null ? 'Unavailable' : `${churnRate}%`}
            </div>
            <div style={{ fontSize: '11px', color: 'var(--hub-text-muted)', marginTop: '2px' }}>
              customers at risk
            </div>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.3 }}
            onClick={() => topSegment && navigate({ to: `/segments/${topSegment.id}` })}
            role={topSegment ? 'button' : undefined}
            tabIndex={topSegment ? 0 : undefined}
            onKeyDown={(event) => {
              if (topSegment) {
                activateOnKeyboard(event, () => navigate({ to: `/segments/${topSegment.id}` }));
              }
            }}
            style={{
              background: 'var(--hub-card)',
              border: '1px solid var(--hub-border)',
              borderRadius: '12px',
              padding: '16px',
              cursor: topSegment ? 'pointer' : 'default',
            }}
          >
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                marginBottom: '8px',
              }}
            >
              <span
                style={{
                  fontSize: '11px',
                  color: 'var(--hub-text-dim)',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                  fontWeight: 700,
                }}
              >
                Top Segment
              </span>
              <div
                style={{
                  width: '28px',
                  height: '28px',
                  borderRadius: '8px',
                  background: 'rgba(232,121,249,0.12)',
                  color: '#e879f9',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                }}
              >
                <PieChart size={13} />
              </div>
            </div>
            <div
              style={{
                fontSize: '16px',
                fontWeight: 700,
                color: 'var(--hub-text)',
                marginTop: '2px',
              }}
            >
              {segmentsLoading ? 'Loading…' : (topSegment?.name ?? '-')}
            </div>
            <div style={{ fontSize: '11px', color: 'var(--hub-text-muted)' }}>
              {segmentsLoading
                ? 'Loading segment data…'
                : segmentsError
                  ? 'Segment data unavailable'
                  : topSegment
                    ? `${topSegment.customerCount.toLocaleString()} members`
                    : 'No segments yet'}
            </div>
          </motion.div>
        </div>

        {/* At-risk customers widget */}
        <motion.div
          initial={{ opacity: 0, y: 8 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.36 }}
          style={{
            background: 'var(--hub-card)',
            border: '1px solid var(--hub-border)',
            borderRadius: '12px',
            overflow: 'hidden',
          }}
        >
          <div
            style={{
              padding: '14px 20px',
              borderBottom: '1px solid var(--hub-border)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
            }}
          >
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
              <AlertTriangle size={14} style={{ color: '#f59e0b' }} />
              <span style={{ fontSize: '13px', fontWeight: 600, color: 'var(--hub-text)' }}>
                At-risk customers
              </span>
              <span style={{ fontSize: '11px', color: 'var(--hub-text-muted)' }}>
                not visited in 30+ days
              </span>
            </div>
            <button
              type="button"
              onClick={() => navigate({ to: '/campaigns/email', search: { name: 'Win-back' } })}
              style={{
                fontSize: '12px',
                color: '#3b82f6',
                background: 'transparent',
                border: 'none',
                cursor: 'pointer',
                fontWeight: 600,
                display: 'flex',
                alignItems: 'center',
                gap: '4px',
              }}
            >
              Send win-back <ArrowRight size={12} />
            </button>
          </div>
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fill, minmax(240px, 1fr))',
              gap: '10px',
              padding: '14px 20px',
            }}
          >
            {customersLoading ? (
              <div
                role="status"
                style={{
                  gridColumn: '1 / -1',
                  fontSize: '12px',
                  color: 'var(--hub-text-muted)',
                  textAlign: 'center',
                  padding: '10px',
                }}
              >
                Loading customer activity…
              </div>
            ) : customersError ? (
              <div
                role="status"
                style={{
                  gridColumn: '1 / -1',
                  fontSize: '12px',
                  color: 'var(--hub-text-muted)',
                  textAlign: 'center',
                  padding: '10px',
                }}
              >
                Customer activity data is unavailable right now.
              </div>
            ) : (
              atRisk.slice(0, 4).map((c) => (
                <div
                  key={c.id}
                  onClick={() => navigate({ to: `/customers/${c.id}` })}
                  role="button"
                  tabIndex={0}
                  onKeyDown={(event) =>
                    activateOnKeyboard(event, () => navigate({ to: `/customers/${c.id}` }))
                  }
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '10px',
                    padding: '8px',
                    borderRadius: '8px',
                    cursor: 'pointer',
                    background: 'var(--hub-hover)',
                  }}
                >
                  <CustomerAvatar initials={c.initials} points={c.loyaltyPoints} size={30} />
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ fontSize: '12px', fontWeight: 600, color: 'var(--hub-text)' }}>
                      {c.name}
                    </div>
                    <div style={{ fontSize: '10px', color: 'var(--hub-text-muted)' }}>
                      Last: {formatCustomerDate(c.lastVisit)}
                    </div>
                  </div>
                </div>
              ))
            )}
            {!customersLoading && !customersError && atRisk.length === 0 && (
              <div
                style={{
                  fontSize: '12px',
                  color: 'var(--hub-text-muted)',
                  textAlign: 'center',
                  padding: '10px',
                }}
              >
                No at-risk customers right now.
              </div>
            )}
          </div>
        </motion.div>

        <div className="crm-form-grid" style={{ display: 'grid', gap: '20px' }}>
          {/* Top customers */}
          <motion.div
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.28, duration: 0.22, ease: 'easeOut' }}
            style={{
              background: 'var(--hub-card)',
              border: '1px solid var(--hub-border)',
              borderRadius: '12px',
              overflow: 'hidden',
            }}
          >
            <div style={{ padding: '16px 20px', borderBottom: '1px solid var(--hub-border)' }}>
              <span style={{ fontSize: '13px', fontWeight: 600, color: 'var(--hub-text)' }}>
                Top Customers
              </span>
            </div>
            <div>
              {dashboardError ? (
                <div
                  role="status"
                  style={{ padding: '20px', color: 'var(--hub-text-muted)', fontSize: '13px' }}
                >
                  Customer ranking data is unavailable right now.
                </div>
              ) : topCustomers.length === 0 ? (
                <div style={{ padding: '20px', color: 'var(--hub-text-muted)', fontSize: '13px' }}>
                  No customer ranking data yet.
                </div>
              ) : (
                topCustomers.map((c, i) => {
                  const color = tierColor[c.tier] ?? '#94a3b8';
                  return (
                    <div
                      key={[c.name, i].join('-')}
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: '12px',
                        padding: '12px 20px',
                        borderBottom:
                          i < topCustomers.length - 1 ? '1px solid var(--hub-border)' : 'none',
                      }}
                    >
                      {/* Avatar */}
                      <div
                        style={{
                          width: '32px',
                          height: '32px',
                          borderRadius: '8px',
                          background: `rgba(59,130,246,0.12)`,
                          color: ACCENT,
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          fontSize: '12px',
                          fontWeight: 700,
                          flexShrink: 0,
                        }}
                      >
                        {c.name
                          .split(' ')
                          .map((n) => n[0])
                          .join('')}
                      </div>
                      <div style={{ flex: 1, minWidth: 0 }}>
                        <div
                          style={{ fontSize: '13px', fontWeight: 500, color: 'var(--hub-text)' }}
                        >
                          {c.name}
                        </div>
                        <div style={{ fontSize: '11px', color: 'var(--hub-text-muted)' }}>
                          {c.visits === undefined ? 'Visits unavailable' : `${c.visits} visits`}
                        </div>
                      </div>
                      <div style={{ textAlign: 'right' }}>
                        <div
                          style={{ fontSize: '13px', fontWeight: 600, color: 'var(--hub-text)' }}
                        >
                          {c.spend}
                        </div>
                        <span
                          style={{
                            display: 'inline-block',
                            padding: '1px 7px',
                            borderRadius: '9999px',
                            fontSize: '10px',
                            fontWeight: 600,
                            background: `${color}18`,
                            color,
                          }}
                        >
                          {c.tier || 'Unranked'}
                        </span>
                      </div>
                    </div>
                  );
                })
              )}
            </div>
          </motion.div>

          {/* Recent campaign performance */}
          <motion.div
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.34, duration: 0.22, ease: 'easeOut' }}
            style={{
              background: 'var(--hub-card)',
              border: '1px solid var(--hub-border)',
              borderRadius: '12px',
              overflow: 'hidden',
            }}
          >
            <div style={{ padding: '16px 20px', borderBottom: '1px solid var(--hub-border)' }}>
              <span style={{ fontSize: '13px', fontWeight: 600, color: 'var(--hub-text)' }}>
                Recent Campaign Performance
              </span>
            </div>
            <div>
              {dashboardError ? (
                <div
                  role="status"
                  style={{ padding: '20px', color: 'var(--hub-text-muted)', fontSize: '13px' }}
                >
                  Campaign performance data is unavailable right now.
                </div>
              ) : campaigns.length === 0 ? (
                <div
                  style={{ padding: '24px 20px', color: 'var(--hub-text-muted)', fontSize: '12px' }}
                >
                  No recent campaigns.
                </div>
              ) : (
                campaigns.map((c, i) => (
                  <div
                    key={[c.name, i].join('-')}
                    style={{
                      padding: '14px 20px',
                      borderBottom:
                        i < campaigns.length - 1 ? '1px solid var(--hub-border)' : 'none',
                    }}
                  >
                    <div
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'space-between',
                        marginBottom: '8px',
                      }}
                    >
                      <div>
                        <div
                          style={{ fontSize: '13px', fontWeight: 500, color: 'var(--hub-text)' }}
                        >
                          {c.name}
                        </div>
                        <div style={{ fontSize: '11px', color: 'var(--hub-text-muted)' }}>
                          {c.channel} · {c.sent.toLocaleString()} sent
                        </div>
                      </div>
                      <div style={{ textAlign: 'right' }}>
                        <div style={{ fontSize: '13px', fontWeight: 600, color: '#10b981' }}>
                          {c.revenue}
                        </div>
                        <div style={{ fontSize: '11px', color: 'var(--hub-text-muted)' }}>
                          {c.openRate} open
                        </div>
                      </div>
                    </div>
                    {/* Open rate bar */}
                    <div
                      style={{
                        height: '4px',
                        borderRadius: '9999px',
                        background: 'var(--hub-border)',
                      }}
                    >
                      <div
                        style={{
                          height: '100%',
                          width: c.openRate,
                          borderRadius: '9999px',
                          background: ACCENT,
                        }}
                      />
                    </div>
                  </div>
                ))
              )}
            </div>
          </motion.div>
        </div>
      </div>
    </>
  );
};

export default Dashboard;
