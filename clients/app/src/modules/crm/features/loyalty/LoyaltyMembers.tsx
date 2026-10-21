import React, { useEffect, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useNavigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { AlertCircle, Search } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { useLoyaltyMembers } from '../../shared/hooks/queries/useLoyalty';
import type { LoyaltyMember } from '../../shared/hooks/queries/useLoyalty';
import { Button, PageHeader, Card, SkeletonList } from '../../shared/components/ui';
import { CustomerAvatar, TierBadge, PointsBadge } from '../../shared/components/crm';
import { getCustomerTier } from '../../shared/lib/tiers';
import { Users } from 'lucide-react';
import { hasQueryFailed } from '../../../../shared/lib/queryFailureUtils';
import { useVirtualRows } from '../../../../shared/hooks/useVirtualRows';
import { VirtualSpacer } from '../../../../shared/components/VirtualList';

const LoyaltyMembers: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useTranslation();
  const pageSize = 100;
  const [offset, setOffset] = useState(0);
  // Set once a "Load more" page comes back empty.
  const [exhausted, setExhausted] = useState(false);
  const [loadedMembers, setLoadedMembers] = useState<LoyaltyMember[]>([]);
  const membersQuery = useLoyaltyMembers(pageSize, offset);
  const { data: fetched, isFetching, refetch } = membersQuery;
  const isError = hasQueryFailed(membersQuery);
  const isLoading = membersQuery.isPending && !isError;
  const [search, setSearch] = useState('');
  useEffect(() => {
    if (!fetched) return;
    if (offset > 0 && fetched.length === 0) {
      // Exact page-size multiple: the extra page came back empty — stop
      // offering "Load more".
      setExhausted(true);
    }
    setLoadedMembers((previous) => {
      if (offset === 0) return fetched;
      const byId = new Map(previous.map((member) => [member.id, member]));
      fetched.forEach((member) => byId.set(member.id, member));
      return Array.from(byId.values());
    });
  }, [fetched, offset]);

  const members = [...(loadedMembers ?? [])].sort((a, b) => b.points - a.points);
  const filtered = members.filter((m) => m.name.toLowerCase().includes(search.toLowerCase()));
  const virtual = useVirtualRows(filtered.length, { estimateSize: 61 });
  const hasMore = !exhausted && fetched?.length === pageSize;

  return (
    <>
      <HelmetContainer title={`${t('crm.loyaltyMembers.pageTitle')} · rtylr`} />
      <div style={{ maxWidth: '1000px' }}>
        <PageHeader
          title={t('crm.loyaltyMembers.pageTitle')}
          subtitle={`${members.length} active members`}
        />

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
              placeholder={t('crm.loyaltyMembers.searchPlaceholder')}
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
              icon={AlertCircle}
              title={t('crm.loyaltyMembers.loadFailedTitle')}
              description={t('crmErrors.refreshAndTryAgain')}
              action={{ label: t('crm.actions.retry'), onClick: () => refetch() }}
            />
          ) : filtered.length === 0 ? (
            <EmptyState
              icon={Users}
              title={t('crm.loyaltyMembers.empty')}
              description={
                search
                  ? t('crm.loyaltyMembers.searchEmptyDescription')
                  : t('crm.loyaltyMembers.trueEmptyDescription')
              }
              action={
                search
                  ? { label: t('crm.actions.clearSearch'), onClick: () => setSearch('') }
                  : {
                      label: t('crm.actions.viewCustomers'),
                      onClick: () => navigate({ to: '/crm/customers' }),
                    }
              }
            />
          ) : (
            <div ref={virtual.listRef}>
              <VirtualSpacer height={virtual.before} />
              {virtual.indexes.map((i) => {
                const m = filtered[i];
                return (
                  <div
                    key={m.id}
                    {...virtual.rowProps(i)}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: '14px',
                      padding: '12px 16px',
                      borderBottom:
                        i < filtered.length - 1 ? '1px solid var(--hub-border)' : 'none',
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
                    <TierBadge tier={getCustomerTier(m.points, m.tier)} />
                    <PointsBadge points={m.points} trend="up" />
                  </div>
                );
              })}
              <VirtualSpacer height={virtual.after} />
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
