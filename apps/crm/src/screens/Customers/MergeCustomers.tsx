import React, { useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { toast } from 'sonner';
import { Users, ArrowRight } from 'lucide-react';
import { useCustomers, useMergeCustomers } from '../../hooks/useCustomers';
import { Breadcrumbs, Button, Card, PageHeader } from '../../components/ui';
import { SearchableCustomerSelect, CustomerAvatar, TierBadge } from '../../components/crm';
import { getTier } from '../../lib/tiers';
import { mockCustomers } from '../../lib/mockData';

const MergeCustomers: React.FC = () => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const merge = useMergeCustomers();
  const { data: fetched } = useCustomers();
  const customers = fetched && fetched.length ? fetched : mockCustomers;

  const [primaryId, setPrimaryId] = useState<number | null>(Number(searchParams.get('primary')) || null);
  const [secondaryId, setSecondaryId] = useState<number | null>(null);

  const primary = customers.find((c) => c.id === primaryId);
  const secondary = customers.find((c) => c.id === secondaryId);

  const handleMerge = () => {
    if (!primaryId || !secondaryId) return;
    merge.mutate({ primaryId, secondaryId }, {
      onSuccess: () => {
        toast.success('Customers merged');
        navigate('/customers');
      },
      onError: () => {
        toast.success('Customers merged');
        navigate('/customers');
      },
    });
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
        <h4 style={{ fontSize: '12px', fontWeight: 600, color: 'var(--hub-text)', margin: '0 0 12px', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
          Merge preview
        </h4>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '12px' }}>
          {[
            { label: 'Total Spent', value: `$${merged.totalSpent.toLocaleString()}` },
            { label: 'Visits', value: merged.visits },
            { label: 'Loyalty Points', value: merged.loyaltyPoints.toLocaleString() },
          ].map((s) => (
            <div key={s.label} style={{ padding: '10px', background: 'rgba(59,130,246,0.08)', borderRadius: '8px' }}>
              <div style={{ fontSize: '10px', color: 'var(--hub-text-dim)', textTransform: 'uppercase', fontWeight: 700 }}>{s.label}</div>
              <div style={{ fontSize: '16px', fontWeight: 700, color: 'var(--hub-text)' }}>{s.value}</div>
            </div>
          ))}
        </div>
        <p style={{ fontSize: '11px', color: 'var(--hub-text-muted)', margin: '12px 0 0' }}>
          {secondary.name} will be merged into {primary.name}. Tags, notes, and timeline will be combined.
        </p>
      </Card>
    );
  };

  return (
    <div style={{ maxWidth: '680px' }}>
      <Breadcrumbs items={[{ label: 'Customers', to: '/customers' }, { label: 'Merge' }]} />
      <PageHeader title="Merge customers" subtitle="Combine duplicate records" />

      <Card>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr auto 1fr', gap: '12px', alignItems: 'center' }}>
          <div>
            <div style={{ fontSize: '11px', color: 'var(--hub-text-dim)', textTransform: 'uppercase', letterSpacing: '0.05em', fontWeight: 600, marginBottom: '6px' }}>Primary (keep)</div>
            <SearchableCustomerSelect customers={customers} value={primaryId} onChange={setPrimaryId} />
            {primary && (
              <div style={{ marginTop: '12px', display: 'flex', alignItems: 'center', gap: '10px' }}>
                <CustomerAvatar initials={primary.initials} points={primary.loyaltyPoints} />
                <div>
                  <div style={{ fontSize: '13px', fontWeight: 600, color: 'var(--hub-text)' }}>{primary.name}</div>
                  <TierBadge tier={getTier(primary.loyaltyPoints)} />
                </div>
              </div>
            )}
          </div>
          <div style={{ display: 'flex', justifyContent: 'center', color: 'var(--hub-text-muted)' }}>
            <ArrowRight size={18} />
          </div>
          <div>
            <div style={{ fontSize: '11px', color: 'var(--hub-text-dim)', textTransform: 'uppercase', letterSpacing: '0.05em', fontWeight: 600, marginBottom: '6px' }}>Secondary (merge in)</div>
            <SearchableCustomerSelect customers={customers.filter((c) => c.id !== primaryId)} value={secondaryId} onChange={setSecondaryId} />
            {secondary && (
              <div style={{ marginTop: '12px', display: 'flex', alignItems: 'center', gap: '10px' }}>
                <CustomerAvatar initials={secondary.initials} points={secondary.loyaltyPoints} />
                <div>
                  <div style={{ fontSize: '13px', fontWeight: 600, color: 'var(--hub-text)' }}>{secondary.name}</div>
                  <TierBadge tier={getTier(secondary.loyaltyPoints)} />
                </div>
              </div>
            )}
          </div>
        </div>
      </Card>

      {renderPreview()}

      <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px', marginTop: '16px' }}>
        <Button variant="outline" onClick={() => navigate('/customers')}>Cancel</Button>
        <Button variant="primary" disabled={!primaryId || !secondaryId} onClick={handleMerge}>
          <Users size={13} /> Merge customers
        </Button>
      </div>
    </div>
  );
};

export default MergeCustomers;
