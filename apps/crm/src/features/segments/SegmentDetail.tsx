import React from 'react';
import { HelmetContainer } from '@voxire/components';
import { useParams, useNavigate } from '@tanstack/react-router';
import { ArrowLeft, Mail, MessageSquare, PieChart } from 'lucide-react';
import { useGetSegment, useSegmentCustomers } from '../../shared/hooks/queries/useSegments';
import { Breadcrumbs, Button, Card, EmptyState, SkeletonList } from '../../shared/components/ui';
import { CustomerAvatar, TierBadge } from '../../shared/components/crm';
import { getCustomerTier } from '../../shared/lib/tiers';

const SegmentDetail: React.FC = () => {
  const { id } = useParams({ strict: false });
  const navigate = useNavigate();
  const { data: fetched, isLoading: segmentLoading, error: segmentError } = useGetSegment(id);
  const { data: fetchedCustomers, isLoading, isError: customersError } = useSegmentCustomers(id);
  const segment = fetched;
  const customers = fetchedCustomers ?? [];

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
          icon={PieChart}
          title="Segment not found"
          description="This segment could not be loaded."
        />
      </div>
    );
  }

  return (
    <>
      <HelmetContainer title="Segment detail · rtylr" />
      <div style={{ maxWidth: '1100px' }}>
        <Breadcrumbs items={[{ label: 'Segments', to: '/segments' }, { label: segment.name }]} />
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
                onClick={() => navigate({ to: `/campaigns/email?segment=${segment.id}` })}
              >
                <Mail size={13} />
                Email
              </Button>
              <Button
                variant="primary"
                onClick={() => navigate({ to: `/campaigns/sms?segment=${segment.id}` })}
              >
                <MessageSquare size={13} />
                SMS
              </Button>
            </div>
          </div>
        </Card>

        <div
          className="crm-card-grid"
          style={{ display: 'grid', gap: '12px', marginBottom: '20px' }}
        >
          {[
            { label: 'Customers', value: segment.customerCount.toLocaleString() },
            { label: 'Trend', value: segment.trend },
            { label: 'Avg Spend', value: '-' },
          ].map((s) => (
            <Card key={s.label}>
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
              Customers in this segment
            </h4>
          </div>
          {isLoading ? (
            <div style={{ padding: '16px' }}>
              <SkeletonList count={5} />
            </div>
          ) : customersError ? (
            <EmptyState
              icon={PieChart}
              title="Unable to load customers"
              description="Refresh and try again."
            />
          ) : customers.length === 0 ? (
            <EmptyState icon={PieChart} title="No customers" description="This segment is empty." />
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
            </div>
          )}
        </Card>
      </div>
    </>
  );
};

export default SegmentDetail;
