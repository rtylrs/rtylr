import React from 'react';
import { Award, Check } from 'lucide-react';
import { useLoyaltyTiers } from '../../hooks/useLoyalty';
import { PageHeader, Card, SkeletonList } from '../../components/ui';

const LoyaltyTiers: React.FC = () => {
  const { data: fetched, isLoading } = useLoyaltyTiers();
  const tiers = (fetched ?? []).sort((a, b) => a.threshold - b.threshold);

  return (
    <div style={{ maxWidth: '1100px' }}>
      <PageHeader title="Loyalty tiers" subtitle="Configure your tier structure and perks" />

      {isLoading ? (
        <Card><SkeletonList count={4} /></Card>
      ) : (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(240px, 1fr))', gap: '14px' }}>
          {tiers.map((t) => (
            <div
              key={t.id}
              style={{
                background: 'var(--hub-card)',
                border: `1px solid ${t.color}40`,
                borderRadius: '14px',
                padding: '18px',
                position: 'relative',
                overflow: 'hidden',
              }}
            >
              <div
                style={{
                  position: 'absolute',
                  top: 0,
                  left: 0,
                  right: 0,
                  height: '4px',
                  background: t.color,
                }}
              />
              <div
                style={{
                  width: '48px',
                  height: '48px',
                  borderRadius: '12px',
                  background: `${t.color}14`,
                  color: t.color,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  marginBottom: '10px',
                }}
              >
                <Award size={22} />
              </div>
              <h3 style={{ fontSize: '18px', fontWeight: 700, color: 'var(--hub-text)', margin: '0 0 4px' }}>{t.name}</h3>
              <div style={{ fontSize: '11px', color: 'var(--hub-text-dim)', textTransform: 'uppercase', letterSpacing: '0.05em', fontWeight: 600 }}>
                {t.threshold === 0 ? 'Starting tier' : `${t.threshold}+ points`}
              </div>
              <div
                style={{
                  fontSize: '12px',
                  color: 'var(--hub-text)',
                  marginTop: '10px',
                  padding: '6px 10px',
                  background: 'var(--hub-hover)',
                  borderRadius: '8px',
                  display: 'inline-block',
                }}
              >
                {t.multiplier}x points multiplier
              </div>
              <div style={{ marginTop: '14px', display: 'flex', flexDirection: 'column', gap: '6px' }}>
                {t.perks.map((p) => (
                  <div key={p} style={{ display: 'flex', alignItems: 'flex-start', gap: '6px', fontSize: '12px', color: 'var(--hub-text-muted)' }}>
                    <Check size={12} style={{ color: t.color, flexShrink: 0, marginTop: '2px' }} />
                    {p}
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default LoyaltyTiers;
