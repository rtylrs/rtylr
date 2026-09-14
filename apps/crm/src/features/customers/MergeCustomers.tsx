import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { useNavigate, useRouterState } from '@tanstack/react-router';
import { toast } from 'sonner';
import { Users, ArrowRight } from 'lucide-react';
import { useListCustomers, useMergeCustomers } from '../../shared/hooks/queries/useCustomers';
import { Breadcrumbs, Button, Card, PageHeader } from '../../shared/components/ui';
import { SearchableCustomerSelect, CustomerAvatar, TierBadge } from '../../shared/components/crm';
import { getCustomerTier } from '../../shared/lib/tiers';

const MergeCustomers: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const searchStr = useRouterState({ select: (s) => s.location.searchStr });
  const merge = useMergeCustomers();
  // Backend caps page size at 100 — the picker covers the first page.
  const { data: fetched } = useListCustomers({ limit: 100, offset: 0 });
  const customers = fetched ?? [];

  const [primaryId, setPrimaryId] = useState<number | null>(
    Number(new URLSearchParams(searchStr).get('primary')) || null
  );
  const [secondaryId, setSecondaryId] = useState<number | null>(null);

  const primary = customers.find((c) => c.id === primaryId);
  const secondary = customers.find((c) => c.id === secondaryId);

  const handleMerge = () => {
    if (!primaryId || !secondaryId) return;
    if (primaryId === secondaryId) {
      toast.error(t('crm.mergeCustomers.sameCustomer'));
      return;
    }
    merge.mutate(
      { primaryId, secondaryId },
      {
        onSuccess: () => {
          toast.success(t('crm.customers.merged'));
          navigate({ to: '/customers' });
        },
        onError: () => {
          toast.error(t('crm.customers.mergeFailed'));
        },
      }
    );
  };

  const renderPreview = () => {
    if (!primary || !secondary) return null;
    const merged = {
      name: primary.name,
      email: primary.email,
      phone: primary.phone,
      totalSpent: primary.totalSpent + secondary.totalSpent,
      visits: primary.visits + secondary.visits,
      loyaltyPoints: primary.loyaltyPoints + secondary.loyaltyPoints,
    };
    return (
      <Card style={{ marginTop: '16px' }}>
        <h4
          style={{
            fontSize: '12px',
            fontWeight: 600,
            color: 'var(--hub-text)',
            margin: '0 0 12px',
            textTransform: 'uppercase',
            letterSpacing: '0.05em',
          }}
        >
          {t('crm.mergeCustomers.preview')}
        </h4>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '12px' }}>
          {[
            {
              id: 'totalSpent',
              label: t('crm.customers.stats.totalSpent'),
              value: `$${merged.totalSpent.toLocaleString()}`,
            },
            { id: 'visits', label: t('crm.customers.stats.visits'), value: merged.visits },
            {
              id: 'loyaltyPoints',
              label: t('crm.customers.stats.loyaltyPoints'),
              value: merged.loyaltyPoints.toLocaleString(),
            },
          ].map((s) => (
            <div
              key={s.id}
              style={{ padding: '10px', background: 'rgba(59,130,246,0.08)', borderRadius: '8px' }}
            >
              <div
                style={{
                  fontSize: '10px',
                  color: 'var(--hub-text-dim)',
                  textTransform: 'uppercase',
                  fontWeight: 700,
                }}
              >
                {s.label}
              </div>
              <div style={{ fontSize: '16px', fontWeight: 700, color: 'var(--hub-text)' }}>
                {s.value}
              </div>
            </div>
          ))}
        </div>
        <p style={{ fontSize: '11px', color: 'var(--hub-text-muted)', margin: '12px 0 0' }}>
          {secondary.name} will be merged into {primary.name}. Tags, notes, and timeline will be
          combined.
        </p>
      </Card>
    );
  };

  return (
    <>
      <HelmetContainer title={`${t('crm.mergeCustomers.pageTitle')} · rtylr`} />
      <div style={{ maxWidth: '680px' }}>
        <Breadcrumbs
          items={[
            { label: t('nav.customers'), to: '/customers' },
            { label: t('crm.customers.breadcrumbMerge') },
          ]}
        />
        <PageHeader
          title={t('crm.mergeCustomers.title')}
          subtitle={t('crm.mergeCustomers.subtitle')}
        />

        <Card>
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: '1fr auto 1fr',
              gap: '12px',
              alignItems: 'center',
            }}
          >
            <div>
              <div
                style={{
                  fontSize: '11px',
                  color: 'var(--hub-text-dim)',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                  fontWeight: 600,
                  marginBottom: '6px',
                }}
              >
                Primary (keep)
              </div>
              <SearchableCustomerSelect
                customers={customers}
                value={primaryId}
                onChange={setPrimaryId}
              />
              {primary && (
                <div
                  style={{ marginTop: '12px', display: 'flex', alignItems: 'center', gap: '10px' }}
                >
                  <CustomerAvatar initials={primary.initials} points={primary.loyaltyPoints} />
                  <div>
                    <div style={{ fontSize: '13px', fontWeight: 600, color: 'var(--hub-text)' }}>
                      {primary.name}
                    </div>
                    <TierBadge tier={getCustomerTier(primary.loyaltyPoints, primary.loyaltyTier)} />
                  </div>
                </div>
              )}
            </div>
            <div
              style={{ display: 'flex', justifyContent: 'center', color: 'var(--hub-text-muted)' }}
            >
              <ArrowRight size={18} />
            </div>
            <div>
              <div
                style={{
                  fontSize: '11px',
                  color: 'var(--hub-text-dim)',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                  fontWeight: 600,
                  marginBottom: '6px',
                }}
              >
                Secondary (merge in)
              </div>
              <SearchableCustomerSelect
                customers={customers.filter((c) => c.id !== primaryId)}
                value={secondaryId}
                onChange={setSecondaryId}
              />
              {secondary && (
                <div
                  style={{ marginTop: '12px', display: 'flex', alignItems: 'center', gap: '10px' }}
                >
                  <CustomerAvatar initials={secondary.initials} points={secondary.loyaltyPoints} />
                  <div>
                    <div style={{ fontSize: '13px', fontWeight: 600, color: 'var(--hub-text)' }}>
                      {secondary.name}
                    </div>
                    <TierBadge
                      tier={getCustomerTier(secondary.loyaltyPoints, secondary.loyaltyTier)}
                    />
                  </div>
                </div>
              )}
            </div>
          </div>
        </Card>

        {renderPreview()}

        <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px', marginTop: '16px' }}>
          <Button variant="outline" onClick={() => navigate({ to: '/customers' })}>
            {t('common.cancel')}
          </Button>
          <Button
            variant="primary"
            disabled={!primaryId || !secondaryId || merge.isPending}
            onClick={handleMerge}
          >
            <Users size={13} /> {merge.isPending ? 'Merging…' : 'Merge customers'}
          </Button>
        </div>
      </div>
    </>
  );
};

export default MergeCustomers;
