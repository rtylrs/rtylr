import React from 'react';
import { HelmetContainer } from '@voxire/components';
import { useParams, useNavigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { ArrowLeft, FileQuestion, Mail, MessageSquare, PieChart, Users } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { useGetSegment, useSegmentCustomers } from '../../shared/hooks/queries/useSegments';
import { Breadcrumbs, Button, Card, SkeletonList } from '../../shared/components/ui';
import { CustomerAvatar, TierBadge } from '../../shared/components/crm';
import { getCustomerTier } from '../../shared/lib/tiers';

const SegmentDetail: React.FC = () => {
  const { id } = useParams({ from: '/layout/segments/$id' });
  const navigate = useNavigate();
  const { t } = useTranslation();
  const { data: fetched, isLoading: segmentLoading, error: segmentError } = useGetSegment(id);
  const {
    data: fetchedCustomerPages,
    isLoading,
    isFetchingNextPage,
    hasNextPage,
    fetchNextPage,
    refetch: refetchCustomers,
  } = useSegmentCustomers(id);
  const segment = fetched;
  const customers = fetchedCustomerPages?.pages.flatMap((page) => page.customers) ?? [];

  if (segmentLoading) {
    return (
      <div style={{ maxWidth: '1100px' }}>
        <SkeletonList count={6} />
      </div>
    );
  }

  if (segmentError || !segment) {
    return (
      <div style={{ maxWidth: '1100px' }}>
        <EmptyState
          icon={FileQuestion}
          title={t('crm.segmentDetail.notFoundTitle')}
          description={t('crm.segmentDetail.notFoundDescription')}
          action={{
            label: t('crm.segmentDetail.backToSegments'),
            onClick: () => navigate({ to: '/segments' }),
          }}
        />
      </div>
    );
  }

  return (
    <>
      <HelmetContainer title={`${t('crm.segmentDetail.pageTitle')} · rtylr`} />
      <div style={{ maxWidth: '1100px' }}>
        <Breadcrumbs
          items={[{ label: t('nav.segments'), to: '/segments' }, { label: segment.name }]}
        />
        <button
          type="button"
          onClick={() => navigate({ to: '/segments' })}
          style={{
            background: 'transparent',
            border: 'none',
            color: 'var(--hub-text-muted)',
            cursor: 'pointer',
            fontSize: '12px',
            display: 'flex',
            alignItems: 'center',
            gap: '4px',
            padding: 0,
            marginBottom: '14px',
          }}
        >
          <ArrowLeft size={12} /> Back
        </button>

        <Card style={{ marginBottom: '20px' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '14px' }}>
            <div
              style={{
                width: '48px',
                height: '48px',
                borderRadius: '12px',
                background: `${segment.color}14`,
                color: segment.color,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              <PieChart size={20} />
            </div>
            <div style={{ flex: 1 }}>
              <h2
                style={{ fontSize: '20px', fontWeight: 700, color: 'var(--hub-text)', margin: 0 }}
              >
                {segment.name}
              </h2>
              <p style={{ fontSize: '12px', color: 'var(--hub-text-muted)', margin: '2px 0 0' }}>
                {segment.description}
              </p>
              <div style={{ display: 'flex', gap: '6px', marginTop: '8px' }}>
                {segment.rules.map((r) => (
                  <span
                    key={r}
                    style={{
                      padding: '2px 8px',
                      borderRadius: '9999px',
                      fontSize: '10px',
                      background: 'var(--hub-hover)',
                      color: 'var(--hub-text-muted)',
                      border: '1px solid var(--hub-border)',
                      fontWeight: 500,
                    }}
                  >
                    {r}
                  </span>
                ))}
              </div>
            </div>
            <div style={{ display: 'flex', gap: '6px' }}>
              <Button
                variant="outline"
                onClick={() => navigate({ to: '/campaigns/email', search: { segment: id } })}
              >
                <Mail size={13} />
                {t('common.email')}
              </Button>
              <Button
                variant="primary"
                onClick={() => navigate({ to: '/campaigns/sms', search: { segment: id } })}
              >
                <MessageSquare size={13} />
                {t('crm.components.sms')}
              </Button>
            </div>
          </div>
        </Card>

        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(3, 1fr)',
            gap: '12px',
            marginBottom: '20px',
          }}
        >
          {[
            {
              id: 'customers',
              label: t('crm.customers.pageTitle'),
              value: segment.customerCount.toLocaleString(),
            },
            { id: 'trend', label: t('crm.segmentDetail.trend'), value: segment.trend },
            { id: 'avgSpend', label: t('crm.segmentDetail.avgSpend'), value: '-' },
          ].map((s) => (
            <Card key={s.id}>
              <div
                style={{
                  fontSize: '10px',
                  color: 'var(--hub-text-dim)',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                  fontWeight: 700,
                }}
              >
                {s.label}
              </div>
              <div
                style={{
                  fontSize: '22px',
                  fontWeight: 700,
                  color: 'var(--hub-text)',
                  marginTop: '4px',
                }}
              >
                {s.value}
              </div>
            </Card>
          ))}
        </div>

        <Card padded={false}>
          <div style={{ padding: '14px 16px', borderBottom: '1px solid var(--hub-border)' }}>
            <h4
              style={{
                fontSize: '12px',
                fontWeight: 600,
                color: 'var(--hub-text)',
                margin: 0,
                textTransform: 'uppercase',
                letterSpacing: '0.05em',
              }}
            >
              {t('crm.segmentDetail.customersIn')}
            </h4>
          </div>
          {isLoading ? (
            <div style={{ padding: '16px' }}>
              <SkeletonList count={5} />
            </div>
          ) : customers.length === 0 ? (
            <EmptyState
              icon={Users}
              title={t('crm.segmentDetail.noCustomersTitle')}
              description={t('crm.segmentDetail.noCustomersDescription')}
              action={{ label: t('crm.actions.refresh'), onClick: () => refetchCustomers() }}
            />
          ) : (
            <div>
              {customers.map((c, i) => (
                <div
                  key={c.id}
                  onClick={() => navigate({ to: `/customers/${c.id}` })}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '12px',
                    padding: '10px 16px',
                    cursor: 'pointer',
                    borderBottom: i < customers.length - 1 ? '1px solid var(--hub-border)' : 'none',
                  }}
                  onMouseEnter={(e) => {
                    (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)';
                  }}
                  onMouseLeave={(e) => {
                    (e.currentTarget as HTMLElement).style.background = 'transparent';
                  }}
                >
                  <CustomerAvatar initials={c.initials} points={c.loyaltyPoints} size={32} />
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ fontSize: '13px', fontWeight: 500, color: 'var(--hub-text)' }}>
                      {c.name}
                    </div>
                    <div style={{ fontSize: '11px', color: 'var(--hub-text-muted)' }}>
                      {c.email}
                    </div>
                  </div>
                  <TierBadge tier={getCustomerTier(c.loyaltyPoints, c.loyaltyTier)} />
                  <span
                    style={{
                      fontSize: '12px',
                      color: 'var(--hub-text-muted)',
                      minWidth: '80px',
                      textAlign: 'right',
                    }}
                  >
                    ${c.totalSpent.toLocaleString()}
                  </span>
                </div>
              ))}
              {hasNextPage && (
                <div style={{ display: 'flex', justifyContent: 'center', padding: '14px 16px' }}>
                  <Button
                    variant="outline"
                    onClick={() => void fetchNextPage()}
                    disabled={isFetchingNextPage}
                  >
                    {isFetchingNextPage
                      ? t('common.loading')
                      : t('crm.segmentDetail.loadMoreCustomers')}
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

export default SegmentDetail;
