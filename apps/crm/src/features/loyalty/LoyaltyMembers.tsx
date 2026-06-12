import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { Search } from 'lucide-react';
import { useLoyaltyMembers } from '../../shared/hooks/queries/useLoyalty';
import { PageHeader, Card, Input, EmptyState, SkeletonList } from '../../shared/components/ui';
import { CustomerAvatar, TierBadge, PointsBadge } from '../../shared/components/crm';
import { Users } from 'lucide-react';

const LoyaltyMembers: React.FC = () => {
  const { data: fetched, isLoading } = useLoyaltyMembers();
  const [search, setSearch] = useState('');
  const members = (fetched ?? []).sort((a, b) => b.points - a.points);
  const filtered = members.filter((m) => m.name.toLowerCase().includes(search.toLowerCase()));

  return (
    <>
      <HelmetContainer title="Loyalty members · rtylr" />
      <div style={{ maxWidth: '1000px' }}>
        <PageHeader title="Loyalty members" subtitle={`${members.length} active members`} />

        <div style={{ marginBottom: '16px', maxWidth: '320px' }}>
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '8px',
              padding: '0 12px',
              height: '36px',
              borderRadius: '8px',
              background: 'var(--hub-input-bg)',
              border: '1px solid var(--hub-border)',
            }}
          >
            <Search size={14} style={{ color: 'var(--hub-text-muted)' }} />
            <input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search members..."
              style={{
                flex: 1,
                background: 'transparent',
                border: 'none',
                outline: 'none',
                fontSize: '13px',
                color: 'var(--hub-text)',
              }}
            />
          </div>
        </div>

        <Card padded={false}>
          {isLoading ? (
            <div style={{ padding: '16px' }}>
              <SkeletonList count={6} />
            </div>
          ) : filtered.length === 0 ? (
            <EmptyState
              icon={Users}
              title="No members found"
              description="Try a different search term."
            />
          ) : (
            <div>
              {filtered.map((m, i) => (
                <div
                  key={m.id}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '14px',
                    padding: '12px 16px',
                    borderBottom: i < filtered.length - 1 ? '1px solid var(--hub-border)' : 'none',
                  }}
                >
                  <span
                    style={{
                      width: '24px',
                      textAlign: 'right',
                      fontSize: '11px',
                      color: 'var(--hub-text-dim)',
                      fontWeight: 700,
                    }}
                  >
                    {i + 1}
                  </span>
                  <CustomerAvatar initials={m.initials} points={m.points} size={36} />
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ fontSize: '13px', fontWeight: 600, color: 'var(--hub-text)' }}>
                      {m.name}
                    </div>
                    <div style={{ fontSize: '11px', color: 'var(--hub-text-muted)' }}>
                      {m.visits} visits • last active {m.lastActive}
                    </div>
                  </div>
                  <TierBadge tier={m.tier as any} />
                  <PointsBadge points={m.points} trend="up" />
                </div>
              ))}
            </div>
          )}
        </Card>
      </div>
    </>
  );
};

export default LoyaltyMembers;
