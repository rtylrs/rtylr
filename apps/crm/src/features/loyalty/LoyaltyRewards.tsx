import React from 'react';
import { Gift, Package } from 'lucide-react';
import { useLoyaltyRewards } from './useLoyalty';
import { PageHeader, Card, EmptyState, Button, SkeletonList } from '../../shared/components/ui';

const LoyaltyRewards: React.FC = () => {
  const { data: fetched, isLoading } = useLoyaltyRewards();
  const rewards = fetched ?? [];

  return (
    <div style={{ maxWidth: '1100px' }}>
      <PageHeader
        title="Rewards catalog"
        subtitle="What customers can redeem their points for"
        actions={<Button variant="primary">Add reward</Button>}
      />

      {isLoading ? (
        <Card><SkeletonList count={4} /></Card>
      ) : rewards.length === 0 ? (
        <Card><EmptyState icon={Gift} title="No rewards yet" description="Add your first reward to the catalog." /></Card>
      ) : (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(240px, 1fr))', gap: '14px' }}>
          {rewards.map((r) => (
            <div
              key={r.id}
              style={{
                background: 'var(--hub-card)',
                border: '1px solid var(--hub-border)',
                borderRadius: '14px',
                padding: '16px',
                display: 'flex',
                flexDirection: 'column',
                gap: '10px',
                position: 'relative',
              }}
            >
              <div
                style={{
                  width: '44px',
                  height: '44px',
                  borderRadius: '12px',
                  background: 'rgba(232,121,249,0.14)',
                  color: '#e879f9',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                }}
              >
                <Gift size={20} />
              </div>
              <div>
                <div style={{ fontSize: '14px', fontWeight: 700, color: 'var(--hub-text)' }}>{r.name}</div>
                <div style={{ fontSize: '11px', color: 'var(--hub-text-muted)', marginTop: '2px' }}>{r.description}</div>
              </div>
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  paddingTop: '10px',
                  borderTop: '1px solid var(--hub-border)',
                }}
              >
                <span style={{ fontSize: '16px', fontWeight: 700, color: '#3b82f6' }}>{r.pointsCost.toLocaleString()} pts</span>
                <span
                  style={{
                    fontSize: '10px',
                    padding: '2px 8px',
                    borderRadius: '9999px',
                    background: 'var(--hub-hover)',
                    color: 'var(--hub-text-muted)',
                    display: 'inline-flex',
                    alignItems: 'center',
                    gap: '4px',
                  }}
                >
                  <Package size={9} />
                  {r.stock} left
                </span>
              </div>
              <div style={{ fontSize: '10px', color: 'var(--hub-text-dim)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>{r.category}</div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default LoyaltyRewards;
