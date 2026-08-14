import React, { useEffect, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { Search } from 'lucide-react';
import { useLoyaltyMembers } from '../../shared/hooks/queries/useLoyalty';
import type { LoyaltyMember } from '../../shared/hooks/queries/useLoyalty';
import {
  Button,
  PageHeader,
  Card,
  Input,
  EmptyState,
  SkeletonList,
} from '../../shared/components/ui';
import { CustomerAvatar, TierBadge, PointsBadge } from '../../shared/components/crm';
import { Users } from 'lucide-react';

const LoyaltyMembers: React.FC = () => {
  const pageSize = 100;
  const [offset, setOffset] = useState(0);
  const [loadedMembers, setLoadedMembers] = useState<LoyaltyMember[]>([]);
  const { data: fetched, isLoading, isFetching, isError } = useLoyaltyMembers(pageSize, offset);
  const [search, setSearch] = useState('');
  useEffect(() => {
    if (!fetched) return;
    setLoadedMembers((previous) => {
      if (offset === 0) return fetched;
      const byId = new Map(previous.map((member) => [member.id, member]));
      fetched.forEach((member) => byId.set(member.id, member));
      return Array.from(byId.values());
    });
  }, [fetched, offset]);

  const members = [...(loadedMembers ?? [])].sort((a, b) => b.points - a.points);
  const filtered = members.filter((m) => m.name.toLowerCase().includes(search.toLowerCase()));
  const hasMore = fetched?.length === pageSize;

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
          {isLoading && members.length === 0 ? (
            <div style={{ padding: '16px' }}>
              <SkeletonList count={6} />
            </div>
          ) : isError && members.length === 0 ? (
            <EmptyState
              icon={Users}
              title="Unable to load members"
              description="Refresh and try again."
            />
          ) : filtered.length === 0 ? (
            <EmptyState
              icon={Users}
              title="No members found"
              description="Try a different search term."
              action={
                hasMore ? (
                  <Button
                    variant="outline"
                    onClick={() => setOffset((current) => current + pageSize)}
                    disabled={isFetching}
                  >
                    {isFetching ? 'Loading…' : 'Load more members'}
                  </Button>
                ) : undefined
              }
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
              {hasMore && (
                <div style={{ display: 'flex', justifyContent: 'center', padding: '16px' }}>
                  <Button
                    variant="outline"
                    onClick={() => setOffset((current) => current + pageSize)}
                    disabled={isFetching}
                  >
                    {isFetching ? 'Loading…' : 'Load more members'}
                  </Button>
                </div>
              )}
            </div>
          )}
        </Card>
      </div>
    </>
  );
};

export default LoyaltyMembers;
