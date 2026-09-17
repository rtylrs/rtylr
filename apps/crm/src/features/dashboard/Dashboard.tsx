import React from 'react';
import { HelmetContainer } from '@voxire/components';
import { useNavigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
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
import { EmptyState } from '@voxire/ui';
import { useCrmDashboard } from '../../shared/hooks/queries/useCrmDashboard';
import { useListCustomers } from '../../shared/hooks/queries/useCustomers';
import { useListSegments } from '../../shared/hooks/queries/useSegments';
import type { DashboardStats } from '../../shared/schemas/dashboard';
import { CustomerAvatar } from '../../shared/components/crm';
import { isAtRisk } from '../../shared/lib/customerUtils';

const ACCENT = '#3b82f6';
const AT_RISK_DAYS = 60;

interface StatItem {
  /** Stable id, used as the React key: the label is a translation and changes with the language. */
  id: string;
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

function StatCard({ stat, index }: { stat: StatItem; index: number }): React.JSX.Element {
  const { t } = useTranslation();
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
            {t('crm.dashboard.noTrendData')}
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
  const { t } = useTranslation();
  const {
    data: dashboardData,
    isLoading: dashLoading,
    refetch: refetchDashboard,
  } = useCrmDashboard();
  // Backend caps page size at 100 — aggregates cover the first page.
  const { data: fetchedCustomers } = useListCustomers({ limit: 100, offset: 0 });
  const { data: fetchedSegments } = useListSegments();
  const customersPool = fetchedCustomers ?? [];
  const segmentsPool = fetchedSegments ?? [];

  const ltv =
    customersPool.length > 0
      ? Math.round(customersPool.reduce((s, c) => s + c.totalSpent, 0) / customersPool.length)
      : 0;
  const atRiskAll = customersPool.filter((c) => isAtRisk(c, AT_RISK_DAYS));
  const churnRate =
    customersPool.length > 0 ? Math.round((atRiskAll.length / customersPool.length) * 100) : 0;
  // Display list is capped; the rate above uses the full pool.
  const atRisk = atRiskAll.slice(0, 4);
  const topSegment =
    segmentsPool.length > 0
      ? segmentsPool.reduce((a, b) => (a.customerCount > b.customerCount ? a : b), segmentsPool[0])
      : null;

  const stats = dashboardData
    ? [
        {
          id: 'totalCustomers',
          label: t('crm.dashboard.totalCustomers'),
          value: (dashboardData as DashboardStats).totalCustomers.toLocaleString(),
          change: dashboardData.totalCustomersChange,
          up: dashboardData.totalCustomersChange?.startsWith('+'),
          icon: Users,
          color: '#3b82f6',
        },
        {
          id: 'newThisMonth',
          label: t('crm.dashboard.newThisMonth'),
          value: dashboardData.newThisMonth.toLocaleString(),
          change: dashboardData.newThisMonthChange,
          up: dashboardData.newThisMonthChange?.startsWith('+'),
          icon: UserPlus,
          color: '#10b981',
        },
        {
          id: 'activeLoyaltyMembers',
          label: t('crm.dashboard.activeLoyaltyMembers'),
          value: dashboardData.activeLoyaltyMembers.toLocaleString(),
          change: dashboardData.activeLoyaltyMembersChange,
          up: dashboardData.activeLoyaltyMembersChange?.startsWith('+'),
          icon: Gift,
          color: '#f59e0b',
        },
        {
          id: 'campaignReach',
          label: t('crm.dashboard.campaignReach'),
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
      <HelmetContainer title={`${t('crm.dashboard.pageTitle')} · rtylr`} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: '24px', maxWidth: '1280px' }}>
        {/* Stats grid */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '16px' }}>
          {dashLoading
            ? [0, 1, 2, 3].map((i) => (
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
            : stats.map((stat, i) => <StatCard key={stat.id} stat={stat} index={i} />)}
        </div>

        {/* Extended KPI widgets */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '16px' }}>
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
                {t('crm.dashboard.avgLtv')}
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
              ${ltv.toLocaleString()}
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
                {t('crm.dashboard.newThisMonth')}
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
              {dashLoading ? <Skeleton width={60} /> : (dashboardData?.newThisMonth ?? 0)}
            </div>
            <div style={{ fontSize: '11px', color: '#10b981', marginTop: '2px' }}>
              {dashboardData?.newThisMonthChange ?? ''}
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
                {t('crm.dashboard.churnRate')}
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
              {churnRate}%
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
                {t('crm.dashboard.topSegment')}
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
              {topSegment?.name ?? '-'}
            </div>
            <div style={{ fontSize: '11px', color: 'var(--hub-text-muted)' }}>
              {topSegment
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
                {t('crm.dashboard.atRiskCustomers')}
              </span>
              <span style={{ fontSize: '11px', color: 'var(--hub-text-muted)' }}>
                not visited in {AT_RISK_DAYS}+ days
              </span>
            </div>
            <button
              type="button"
              onClick={() => navigate({ to: '/campaigns/email', search: { segment: 'atrisk' } })}
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
            {atRisk.slice(0, 4).map((c) => (
              <div
                key={c.id}
                onClick={() => navigate({ to: `/customers/${c.id}` })}
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
                    Last: {c.lastVisit}
                  </div>
                </div>
              </div>
            ))}
            {atRisk.length === 0 && (
              <EmptyState
                size="compact"
                icon={AlertTriangle}
                title={t('crm.dashboard.atRiskEmptyTitle')}
                action={{
                  label: t('crm.actions.viewCustomers'),
                  onClick: () => navigate({ to: '/customers' }),
                }}
              />
            )}
          </div>
        </motion.div>

        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '20px' }}>
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
                {t('crm.dashboard.topCustomers')}
              </span>
            </div>
            <div>
              {topCustomers.map((c, i) => (
                <div
                  key={`${c.name}-${i}`}
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
                    <div style={{ fontSize: '13px', fontWeight: 500, color: 'var(--hub-text)' }}>
                      {c.name}
                    </div>
                    <div style={{ fontSize: '11px', color: 'var(--hub-text-muted)' }}>
                      {c.visits === undefined ? 'Visits unavailable' : `${c.visits} visits`}
                    </div>
                  </div>
                  <div style={{ textAlign: 'right' }}>
                    <div style={{ fontSize: '13px', fontWeight: 600, color: 'var(--hub-text)' }}>
                      {c.spend}
                    </div>
                    <span
                      style={{
                        display: 'inline-block',
                        padding: '1px 7px',
                        borderRadius: '9999px',
                        fontSize: '10px',
                        fontWeight: 600,
                        background: `${tierColor[c.tier]}18`,
                        color: tierColor[c.tier],
                      }}
                    >
                      {c.tier}
                    </span>
                  </div>
                </div>
              ))}
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
                {t('crm.dashboard.recentCampaignPerformance')}
              </span>
            </div>
            <div>
              {campaigns.map((c, i) => (
                <div
                  key={`${c.name}-${i}`}
                  style={{
                    padding: '14px 20px',
                    borderBottom: i < campaigns.length - 1 ? '1px solid var(--hub-border)' : 'none',
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
                      <div style={{ fontSize: '13px', fontWeight: 500, color: 'var(--hub-text)' }}>
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
              ))}
              {campaigns.length === 0 && (
                <EmptyState
                  size="compact"
                  icon={Megaphone}
                  title={t('crm.dashboard.campaignPerformanceUnavailableTitle')}
                  description={t('crm.dashboard.campaignPerformanceUnavailableDescription')}
                  action={{
                    label: t('crm.actions.refresh'),
                    onClick: () => refetchDashboard(),
                  }}
                />
              )}
            </div>
          </motion.div>
        </div>
      </div>
    </>
  );
};

export default Dashboard;
