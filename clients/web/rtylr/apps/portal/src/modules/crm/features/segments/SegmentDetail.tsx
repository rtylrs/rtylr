import React from 'react';
import { HelmetContainer } from '@voxire/components';
import { useParams, useNavigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import {
  AlertCircle,
  ArrowLeft,
  FileQuestion,
  Mail,
  MessageSquare,
  PieChart,
  Users,
} from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { useGetSegment, useSegmentCustomers } from '../../shared/hooks/queries/useSegments';
import { Breadcrumbs, Button, Card, SkeletonList } from '../../shared/components/ui';
import { CustomerAvatar, TierBadge } from '../../shared/components/crm';
import { getCustomerTier } from '../../shared/lib/tiers';

function errorStatus(error: unknown): number | undefined {
  if (!error || typeof error !== 'object') return undefined;
  const candidate = error as {
    statusCode?: unknown;
    status?: unknown;
    response?: { status?: unknown };
  };
  if (typeof candidate.statusCode === 'number') return candidate.statusCode;
  if (typeof candidate.status === 'number') return candidate.status;
  if (typeof candidate.response?.status === 'number') return candidate.response.status;
  return undefined;
}

function isNotFoundError(error: unknown, notFoundTitle: string): boolean {
  const status = errorStatus(error);
  if (status !== undefined) return status === 404;
  return (
    error instanceof Error && (error.message === notFoundTitle || /not found/i.test(error.message))
  );
}

const LoadErrorState: React.FC<{ onRetry: () => void }> = ({ onRetry }) => {
  const { t } = useTranslation();
  return (
    <EmptyState
      icon={AlertCircle}
      title={t('crmErrors.generic')}
      description={t('crmErrors.refreshAndTryAgain')}
      action={{ label: t('crm.actions.retry'), onClick: onRetry }}
    />
  );
};

const SegmentDetail: React.FC = () => {
  const { id } = useParams({ from: '/business/crm/segments/$id' });
  const navigate = useNavigate();
  const { t } = useTranslation();
  const {
    data: fetched,
    isLoading: segmentLoading,
    error: segmentError,
    refetch: refetchSegment,
  } = useGetSegment(id);
  const {
    data: fetchedCustomerPages,
    isLoading,
    isFetchingNextPage,
    isError: customersError,
    isFetchNextPageError: customersNextPageError,
    hasNextPage,
    fetchNextPage,
    refetch: refetchCustomers,
  } = useSegmentCustomers(id);
  const segment = fetched;
  const customers = fetchedCustomerPages?.pages.flatMap((page) => page.customers) ?? [];
  const customerTotal = fetchedCustomerPages?.pages[0]?.total;
  const customersQueryFailed = customersError || customersNextPageError;
  const notFoundTitle = t('crm.segmentDetail.notFoundTitle');
  const segmentIsMissing = segmentError ? isNotFoundError(segmentError, notFoundTitle) : !segment;
  const retryCustomers = (): void => {
    void (customersNextPageError ? fetchNextPage() : refetchCustomers());
  };

  if (segmentLoading) {
    return (
      <div style={{ maxWidth: '1100px' }}>
        <SkeletonList count={6} />
      </div>
    );
  }

  if (segmentError && !segmentIsMissing) {
    return (
      <div style={{ maxWidth: '1100px' }}>
        <LoadErrorState onRetry={() => void refetchSegment()} />
      </div>
    );
  }

  if (!segment || segmentIsMissing) {
    return (
      <div style={{ maxWidth: '1100px' }}>
        <EmptyState
          icon={FileQuestion}
          title={notFoundTitle}
          description={t('crm.segmentDetail.notFoundDescription')}
          action={{
            label: t('crm.segmentDetail.backToSegments'),
            onClick: () => navigate({ to: '/crm/segments' }),
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
          items={[{ label: t('crmNav.segments'), to: '/crm/segments' }, { label: segment.name }]}
        />
        <button
          type="button"
          onClick={() => navigate({ to: '/crm/segments' })}
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
                onClick={() => navigate({ to: '/crm/campaigns/email', search: { segment: id } })}
              >
                <Mail size={13} />
                {t('crmCommon.email')}
              </Button>
              <Button
                variant="primary"
                onClick={() => navigate({ to: '/crm/campaigns/sms', search: { segment: id } })}
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
              value:
                customerTotal !== undefined
                  ? customerTotal.toLocaleString()
                  : customersQueryFailed
                    ? '—'
                    : t('crmCommon.loading'),
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
          {isLoading && !fetchedCustomerPages ? (
            <div style={{ padding: '16px' }}>
              <SkeletonList count={5} />
            </div>
          ) : customersQueryFailed && customers.length === 0 ? (
            <LoadErrorState onRetry={retryCustomers} />
          ) : customers.length === 0 ? (
            <EmptyState
              icon={Users}
              title={t('crm.segmentDetail.noCustomersTitle')}
              description={t('crm.segmentDetail.noCustomersDescription')}
              action={{ label: t('crm.actions.refresh'), onClick: () => refetchCustomers() }}
            />
          ) : (
            <div>
              {customersQueryFailed && (
                <div
                  role="alert"
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    gap: '12px',
                    padding: '12px 16px',
                    borderBottom: '1px solid var(--hub-border)',
                    color: 'var(--hub-text-muted)',
                    fontSize: '12px',
                  }}
                >
                  <span>{t('crmErrors.refreshAndTryAgain')}</span>
                  <Button variant="outline" onClick={retryCustomers}>
                    {t('crm.actions.retry')}
                  </Button>
                </div>
              )}
              {customers.map((c, i) => (
                <div
                  key={c.id}
                  onClick={() => navigate({ to: `/crm/customers/${c.id}` })}
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
                      ? t('crmCommon.loading')
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
