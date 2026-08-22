import React, { useState, useEffect } from 'react';
import { HelmetContainer } from '@voxire/components';
import {
  useLoyaltyConfig,
  useLoyaltyTransactions,
  useLoyaltyTiers,
  useUpdateLoyaltyConfig,
} from '../../shared/hooks/queries/useLoyalty';
import { motion } from 'motion/react';
import {
  Gift,
  Star,
  Trophy,
  Crown,
  Award,
  ToggleLeft,
  ToggleRight,
  Clock,
  AlertTriangle,
} from 'lucide-react';
import { type LoyaltyTransaction, type LoyaltyTier } from '../../shared/schemas/loyalty';
import { Card, EmptyState, SkeletonList } from '../../shared/components/ui';

const ACCENT = '#3b82f6';

const TIER_ICON_MAP: Record<string, React.ComponentType<{ size?: number; color?: string }>> = {
  Bronze: Award,
  Silver: Star,
  Gold: Trophy,
  Platinum: Crown,
};

const activityColor: Record<string, string> = {
  earn: '#10b981',
  redeem: ACCENT,
  upgrade: '#f59e0b',
  award: '#10b981',
  adjust: '#8b5cf6',
  expire: '#ef4444',
};

const Loyalty: React.FC = () => {
  const { data: config, isLoading: configLoading, isError: configError } = useLoyaltyConfig();
  const {
    data: transactions,
    isLoading: transactionsLoading,
    isError: transactionsError,
  } = useLoyaltyTransactions();
  const { data: tiersData, isLoading: tiersLoading, isError: tiersError } = useLoyaltyTiers();
  const updateConfig = useUpdateLoyaltyConfig();
  const configUnavailable = configLoading || configError;

  const tiers: LoyaltyTier[] = tiersData ?? [];

  const [programEnabled, setProgramEnabled] = useState(true);
  const [pointsPerDollar, setPointsPerDollar] = useState(1);
  const [redemptionRate, setRedemptionRate] = useState(0.01);

  useEffect(() => {
    if (config) {
      if (config.programEnabled !== undefined) setProgramEnabled(config.programEnabled);
      if (config.pointsPerDollar !== undefined) setPointsPerDollar(config.pointsPerDollar);
      if (config.redemptionRate !== undefined) setRedemptionRate(config.redemptionRate);
    }
  }, [config]);

  const activityLog: LoyaltyTransaction[] = transactions ?? [];

  return (
    <>
      <HelmetContainer title="Loyalty · rtylr" />
      <div style={{ display: 'flex', flexDirection: 'column', gap: '24px', maxWidth: '1280px' }}>
        {/* Header */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            flexWrap: 'wrap',
            gap: '12px',
          }}
        >
          <div>
            <h2 style={{ fontSize: '18px', fontWeight: 700, color: 'var(--hub-text)', margin: 0 }}>
              Loyalty Program
            </h2>
            <p style={{ fontSize: '13px', color: 'var(--hub-text-muted)', margin: '2px 0 0' }}>
              Configure tiers, point rules, and member rewards
            </p>
          </div>
          <button
            type="button"
            onClick={() => {
              const next = !programEnabled;
              updateConfig.mutate(
                { programEnabled: next },
                {
                  onSuccess: () => setProgramEnabled(next),
                  onError: () => setProgramEnabled(!next),
                }
              );
            }}
            disabled={updateConfig.isPending || configUnavailable}
            aria-label={programEnabled ? 'Pause loyalty program' : 'Activate loyalty program'}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '8px',
              padding: '8px 16px',
              borderRadius: '8px',
              fontSize: '13px',
              fontWeight: 600,
              background: programEnabled ? 'rgba(16,185,129,0.12)' : 'var(--hub-card)',
              color: programEnabled ? '#10b981' : 'var(--hub-text-muted)',
              border: `1px solid ${programEnabled ? 'rgba(16,185,129,0.3)' : 'var(--hub-border)'}`,
              cursor: updateConfig.isPending || configUnavailable ? 'not-allowed' : 'pointer',
              transition: 'all 0.2s',
              opacity: updateConfig.isPending || configUnavailable ? 0.65 : 1,
            }}
          >
            {programEnabled ? <ToggleRight size={16} /> : <ToggleLeft size={16} />}
            {programEnabled ? 'Program Active' : 'Program Paused'}
          </button>
        </div>

        {/* Config section */}
        <motion.div
          initial={{ opacity: 0, y: 8 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.2 }}
          style={{
            background: 'var(--hub-card)',
            border: '1px solid var(--hub-border)',
            borderRadius: '12px',
            padding: '20px',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '20px' }}>
            <Gift size={16} style={{ color: ACCENT }} />
            <span style={{ fontSize: '14px', fontWeight: 600, color: 'var(--hub-text)' }}>
              Program Configuration
            </span>
          </div>
          <div className="crm-card-grid" style={{ display: 'grid', gap: '20px' }}>
            <div>
              <label
                style={{
                  display: 'block',
                  fontSize: '11px',
                  fontWeight: 600,
                  color: 'var(--hub-text-muted)',
                  marginBottom: '8px',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                }}
              >
                Points per $1 spent
              </label>
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                <button
                  type="button"
                  aria-label="Decrease points per dollar"
                  onClick={() => setPointsPerDollar((v) => Math.max(1, v - 1))}
                  style={{
                    width: '32px',
                    height: '32px',
                    borderRadius: '8px',
                    border: '1px solid var(--hub-border)',
                    background: 'var(--hub-hover)',
                    color: 'var(--hub-text)',
                    cursor: 'pointer',
                    fontSize: '16px',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                  }}
                >
                  −
                </button>
                <span
                  style={{
                    fontSize: '20px',
                    fontWeight: 700,
                    color: 'var(--hub-text)',
                    minWidth: '40px',
                    textAlign: 'center',
                  }}
                >
                  {pointsPerDollar}
                </span>
                <button
                  type="button"
                  aria-label="Increase points per dollar"
                  onClick={() => setPointsPerDollar((v) => Math.min(10, v + 1))}
                  style={{
                    width: '32px',
                    height: '32px',
                    borderRadius: '8px',
                    border: '1px solid var(--hub-border)',
                    background: 'var(--hub-hover)',
                    color: 'var(--hub-text)',
                    cursor: 'pointer',
                    fontSize: '16px',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                  }}
                >
                  +
                </button>
              </div>
            </div>
            <div>
              <label
                style={{
                  display: 'block',
                  fontSize: '11px',
                  fontWeight: 600,
                  color: 'var(--hub-text-muted)',
                  marginBottom: '8px',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                }}
              >
                Points expiry
              </label>
              <div
                style={{
                  height: '36px',
                  padding: '0 12px',
                  borderRadius: '8px',
                  border: '1px solid var(--hub-border)',
                  background: 'var(--hub-input-bg)',
                  color: 'var(--hub-text-muted)',
                  fontSize: '13px',
                  display: 'flex',
                  alignItems: 'center',
                }}
              >
                Not exposed by the loyalty service
              </div>
            </div>
            <div>
              <label
                style={{
                  display: 'block',
                  fontSize: '11px',
                  fontWeight: 600,
                  color: 'var(--hub-text-muted)',
                  marginBottom: '8px',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                }}
              >
                Redemption rate
              </label>
              <div style={{ fontSize: '13px', color: 'var(--hub-text)', lineHeight: '36px' }}>
                {redemptionRate > 0
                  ? `${Math.round(1 / redemptionRate).toLocaleString()} pts = $1.00`
                  : 'Not configured'}
              </div>
            </div>
          </div>
          {configError && (
            <p style={{ color: '#ef4444', fontSize: '12px', margin: '12px 0 0' }}>
              Unable to load the current loyalty configuration. Refresh before saving changes.
            </p>
          )}
          <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: '16px' }}>
            <button
              type="button"
              onClick={() =>
                updateConfig.mutate(
                  { pointsPerDollar },
                  {
                    onError: () => setPointsPerDollar(config?.pointsPerDollar ?? 1),
                  }
                )
              }
              disabled={updateConfig.isPending || configUnavailable}
              style={{
                padding: '8px 20px',
                borderRadius: '8px',
                fontSize: '13px',
                fontWeight: 600,
                background: ACCENT,
                color: '#fff',
                border: 'none',
                cursor: updateConfig.isPending || configUnavailable ? 'not-allowed' : 'pointer',
                opacity: updateConfig.isPending || configUnavailable ? 0.7 : 1,
              }}
            >
              {updateConfig.isPending ? 'Saving…' : 'Save'}
            </button>
          </div>
        </motion.div>

        {/* Tier cards */}
        <div>
          <h3
            style={{
              fontSize: '14px',
              fontWeight: 600,
              color: 'var(--hub-text)',
              marginBottom: '12px',
              marginTop: 0,
            }}
          >
            Loyalty Tiers
          </h3>
          {tiersLoading ? (
            <Card>
              <SkeletonList count={4} />
            </Card>
          ) : tiersError ? (
            <Card>
              <EmptyState
                icon={AlertTriangle}
                title="Unable to load loyalty tiers"
                description="Refresh and try again."
              />
            </Card>
          ) : tiers.length === 0 ? (
            <Card>
              <EmptyState
                icon={Award}
                title="No loyalty tiers configured"
                description="Configure tiers to start rewarding customers."
              />
            </Card>
          ) : (
            <div className="crm-stats-grid" style={{ display: 'grid', gap: '16px' }}>
              {tiers.map((tier, i) => {
                const Icon = TIER_ICON_MAP[tier.name] ?? Award;
                return (
                  <motion.div
                    key={`${tier.id}-${tier.name}-${i}`}
                    initial={{ opacity: 0, y: 10 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ delay: i * 0.06, duration: 0.2 }}
                    style={{
                      background: 'var(--hub-card)',
                      border: `1px solid ${tier.color}30`,
                      borderRadius: '12px',
                      padding: '20px',
                      display: 'flex',
                      flexDirection: 'column',
                      gap: '12px',
                    }}
                  >
                    <div
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'space-between',
                      }}
                    >
                      <div
                        style={{
                          width: '36px',
                          height: '36px',
                          borderRadius: '10px',
                          background: `${tier.color}18`,
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                        }}
                      >
                        <Icon size={18} color={tier.color} />
                      </div>
                    </div>
                    <div>
                      <div style={{ fontSize: '16px', fontWeight: 700, color: 'var(--hub-text)' }}>
                        {tier.name}
                      </div>
                      <div
                        style={{
                          fontSize: '12px',
                          color: 'var(--hub-text-muted)',
                          marginTop: '2px',
                        }}
                      >
                        {tier.threshold === 0
                          ? 'Starting tier'
                          : `From ${tier.threshold.toLocaleString()} pts`}
                      </div>
                    </div>
                    <div
                      style={{
                        padding: '8px 12px',
                        borderRadius: '8px',
                        background: `${tier.color}10`,
                        border: `1px solid ${tier.color}25`,
                      }}
                    >
                      <span style={{ fontSize: '22px', fontWeight: 700, color: tier.color }}>
                        {tier.multiplier}%
                      </span>
                      <span
                        style={{
                          fontSize: '12px',
                          color: 'var(--hub-text-muted)',
                          marginLeft: '4px',
                        }}
                      >
                        discount
                      </span>
                    </div>
                    <ul
                      style={{
                        margin: 0,
                        padding: '0 0 0 16px',
                        display: 'flex',
                        flexDirection: 'column',
                        gap: '4px',
                      }}
                    >
                      {tier.perks.map((perk, perkIndex) => (
                        <li
                          key={`${perk}-${perkIndex}`}
                          style={{ fontSize: '11px', color: 'var(--hub-text-muted)' }}
                        >
                          {perk}
                        </li>
                      ))}
                    </ul>
                  </motion.div>
                );
              })}
            </div>
          )}
        </div>

        {/* Activity log */}
        <motion.div
          initial={{ opacity: 0, y: 8 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.3, duration: 0.2 }}
          style={{
            background: 'var(--hub-card)',
            border: '1px solid var(--hub-border)',
            borderRadius: '12px',
            overflow: 'hidden',
          }}
        >
          <div
            style={{
              padding: '16px 20px',
              borderBottom: '1px solid var(--hub-border)',
              display: 'flex',
              alignItems: 'center',
              gap: '8px',
            }}
          >
            <Clock size={14} style={{ color: ACCENT }} />
            <span style={{ fontSize: '13px', fontWeight: 600, color: 'var(--hub-text)' }}>
              Points Activity Log
            </span>
          </div>
          {transactionsLoading ? (
            <div style={{ padding: '16px' }}>
              <SkeletonList count={3} />
            </div>
          ) : transactionsError ? (
            <div style={{ padding: '20px', color: 'var(--hub-text-muted)', fontSize: '13px' }}>
              Unable to load loyalty activity. Please refresh and try again.
            </div>
          ) : activityLog.length === 0 ? (
            <div style={{ padding: '20px', color: 'var(--hub-text-muted)', fontSize: '13px' }}>
              No loyalty activity yet.
            </div>
          ) : (
            activityLog.map((entry, i) => (
              <div
                key={
                  'id' in entry ? entry.id : `${entry.customer}-${entry.time}-${entry.action}-${i}`
                }
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: '12px',
                  padding: '12px 20px',
                  borderBottom: i < activityLog.length - 1 ? '1px solid var(--hub-border)' : 'none',
                }}
              >
                <div
                  style={{
                    width: '8px',
                    height: '8px',
                    borderRadius: '50%',
                    background: activityColor[entry.type] ?? 'var(--hub-text-dim)',
                    flexShrink: 0,
                  }}
                />
                <div style={{ flex: 1 }}>
                  <span style={{ fontSize: '13px', fontWeight: 500, color: 'var(--hub-text)' }}>
                    {entry.customer}
                  </span>
                  <span style={{ fontSize: '13px', color: 'var(--hub-text-muted)' }}>
                    : {entry.action}
                  </span>
                </div>
                <span style={{ fontSize: '11px', color: 'var(--hub-text-dim)' }}>{entry.time}</span>
              </div>
            ))
          )}
        </motion.div>
      </div>
    </>
  );
};

export default Loyalty;
