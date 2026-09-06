import React, { useState, useEffect } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useUserStore } from '@voxire/auth';
import {
  useLoyaltyConfig,
  useLoyaltyTransactions,
  useLoyaltyTiers,
  useUpdateLoyaltyConfig,
} from '../../shared/hooks/queries/useLoyalty';
import { motion } from 'motion/react';
import { Gift, Star, Trophy, Crown, Award, ToggleLeft, ToggleRight, Clock } from 'lucide-react';
import { type LoyaltyTransaction, type LoyaltyTier } from '../../shared/schemas/loyalty';
import { canCrmPermission } from '../../shared/lib/crmPermissions';

const ACCENT = '#3b82f6';

const TIER_ICON_MAP: Record<string, React.ComponentType<{ size?: number; style?: React.CSSProperties }>> = {
  Bronze: Award,
  Silver: Star,
  Gold: Trophy,
  Platinum: Crown,
};

const activityColor: Record<string, string> = {
  earn: '#10b981',
  award: '#10b981',
  redeem: ACCENT,
  reward: ACCENT,
  upgrade: '#f59e0b',
  adjust: '#8b5cf6',
  expire: '#94a3b8',
};

const Loyalty: React.FC = () => {
  const role = useUserStore((state) => state.user?.role ?? state.role);
  const canConfigureLoyalty = canCrmPermission(role, 'crm:loyalty:configure');
  const { data: config } = useLoyaltyConfig();
  const { data: transactions } = useLoyaltyTransactions();
  const { data: tiersData } = useLoyaltyTiers();
  const updateConfig = useUpdateLoyaltyConfig();

  const tiers: LoyaltyTier[] = tiersData ?? [];

  // Start from the loaded config, not an assumed-on default that flashes the
  // wrong program state before the query resolves.
  const [programEnabled, setProgramEnabled] = useState(false);
  const [pointsPerDollar, setPointsPerDollar] = useState(1);
  const [pointsExpiry, setPointsExpiry] = useState(365);

  useEffect(() => {
    if (config) {
      if (config.programEnabled !== undefined) setProgramEnabled(config.programEnabled);
      if (config.pointsPerDollar !== undefined) setPointsPerDollar(config.pointsPerDollar);
      if (config.pointsExpiry !== undefined) setPointsExpiry(config.pointsExpiry);
    }
  }, [config]);

  const activityLog: LoyaltyTransaction[] = transactions ?? [];

  return (
    <>
      <HelmetContainer title="Loyalty · rtylr" />
      <div style={{ display: 'flex', flexDirection: 'column', gap: '24px', maxWidth: '1280px' }}>
        {/* Header */}
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <div>
            <h2 style={{ fontSize: '18px', fontWeight: 700, color: 'var(--hub-text)', margin: 0 }}>
              Loyalty Program
            </h2>
            <p style={{ fontSize: '13px', color: 'var(--hub-text-muted)', margin: '2px 0 0' }}>
              Configure tiers, point rules, and member rewards
            </p>
          </div>
          {canConfigureLoyalty && (
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
                cursor: 'pointer',
                transition: 'all 0.2s',
              }}
            >
              {programEnabled ? <ToggleRight size={16} /> : <ToggleLeft size={16} />}
              {programEnabled ? 'Program Active' : 'Program Paused'}
            </button>
          )}
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
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '20px' }}>
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
                Points expiry (days)
              </label>
              <select
                value={pointsExpiry}
                onChange={(e) => setPointsExpiry(Number(e.target.value))}
                style={{
                  height: '36px',
                  padding: '0 12px',
                  borderRadius: '8px',
                  border: '1px solid var(--hub-border)',
                  background: 'var(--hub-input-bg)',
                  color: 'var(--hub-text)',
                  fontSize: '13px',
                  cursor: 'pointer',
                  outline: 'none',
                }}
              >
                <option value={90}>90 days</option>
                <option value={180}>180 days</option>
                <option value={365}>1 year</option>
                <option value={730}>2 years</option>
                <option value={0}>Never</option>
              </select>
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
                {config && config.redemptionRate > 0
                  ? `${Math.round(1 / config.redemptionRate)} pts = $1.00`
                  : 'Set a redemption rate'}
              </div>
            </div>
          </div>
          {canConfigureLoyalty && (
            <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: '16px' }}>
              <button
                type="button"
                onClick={() => updateConfig.mutate({ pointsPerDollar, pointsExpiry })}
                disabled={updateConfig.isPending}
                style={{
                  padding: '8px 20px',
                  borderRadius: '8px',
                  fontSize: '13px',
                  fontWeight: 600,
                  background: ACCENT,
                  color: '#fff',
                  border: 'none',
                  cursor: updateConfig.isPending ? 'not-allowed' : 'pointer',
                  opacity: updateConfig.isPending ? 0.7 : 1,
                }}
              >
                {updateConfig.isPending ? 'Saving…' : 'Save'}
              </button>
            </div>
          )}
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
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '16px' }}>
            {tiers.map((tier, i) => {
              const Icon = TIER_ICON_MAP[tier.name] ?? Award;
              return (
                <motion.div
                  key={tier.name}
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
                      <Icon size={18} style={{ color: tier.color }} />
                    </div>
                  </div>
                  <div>
                    <div style={{ fontSize: '16px', fontWeight: 700, color: 'var(--hub-text)' }}>
                      {tier.name}
                    </div>
                    <div
                      style={{ fontSize: '12px', color: 'var(--hub-text-muted)', marginTop: '2px' }}
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
                    {tier.perks.map((perk) => (
                      <li key={perk} style={{ fontSize: '11px', color: 'var(--hub-text-muted)' }}>
                        {perk}
                      </li>
                    ))}
                  </ul>
                </motion.div>
              );
            })}
          </div>
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
          {activityLog.map((entry, i) => (
            <div
              key={entry.id}
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
                  background: activityColor[entry.type] ?? 'var(--hub-border)',
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
          ))}
        </motion.div>
      </div>
    </>
  );
};

export default Loyalty;
