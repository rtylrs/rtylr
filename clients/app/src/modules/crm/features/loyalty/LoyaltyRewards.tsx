import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { useUserStore } from '../../../../shared/stores/userStore';
import { Gift, Package, Plus, X } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { useLoyaltyRewards, useUpsertReward } from '../../shared/hooks/queries/useLoyalty';
import { PageHeader, Card, Button, SkeletonList } from '../../shared/components/ui';
import { canCrmPermission } from '../../shared/lib/crmPermissions';

interface RewardForm {
  name: string;
  description: string;
  pointsCost: number;
  stock: number;
  category: string;
}

const AddRewardModal: React.FC<{ onClose: () => void }> = ({ onClose }) => {
  const { t } = useTranslation();
  const upsert = useUpsertReward();
  const [form, setForm] = useState<RewardForm>({
    name: '',
    description: '',
    pointsCost: 100,
    stock: 10,
    category: 'general',
  });

  const handleSubmit = (): void => {
    if (!form.name.trim()) {
      toast.error(t('crm.rewards.nameRequired'));
      return;
    }
    upsert.mutate(
      { data: { ...form, active: true } },
      {
        onSuccess: () => onClose(),
        onError: () => toast.error(t('crm.rewards.saveFailed')),
      }
    );
  };

  const inputStyle: React.CSSProperties = {
    padding: '8px 12px',
    borderRadius: '8px',
    border: '1px solid var(--hub-border)',
    background: 'var(--hub-input-bg, var(--hub-card))',
    color: 'var(--hub-text)',
    fontSize: '13px',
    outline: 'none',
    width: '100%',
    boxSizing: 'border-box',
  };

  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        background: 'rgba(0,0,0,0.5)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: 9999,
      }}
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div
        style={{
          background: 'var(--hub-card)',
          border: '1px solid var(--hub-border)',
          borderRadius: '16px',
          padding: '24px',
          width: '420px',
          display: 'flex',
          flexDirection: 'column',
          gap: '12px',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <h3 style={{ fontSize: '15px', fontWeight: 700, color: 'var(--hub-text)', margin: 0 }}>
            {t('crm.rewards.newReward')}
          </h3>
          <button
            type="button"
            aria-label={t('crmCommon.close')}
            onClick={onClose}
            style={{
              background: 'none',
              border: 'none',
              cursor: 'pointer',
              color: 'var(--hub-text-muted)',
            }}
          >
            <X size={16} />
          </button>
        </div>
        <input
          placeholder={t('crm.rewards.nameLabel')}
          value={form.name}
          onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
          style={inputStyle}
        />
        <input
          placeholder={t('crm.campaigns.descriptionLabel')}
          value={form.description}
          onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
          style={inputStyle}
        />
        <input
          placeholder={t('crm.rewards.categoryLabel')}
          value={form.category}
          onChange={(e) => setForm((f) => ({ ...f, category: e.target.value }))}
          style={inputStyle}
        />
        <div style={{ display: 'flex', gap: '8px' }}>
          <div style={{ flex: 1 }}>
            <label
              style={{
                fontSize: '11px',
                color: 'var(--hub-text-muted)',
                display: 'block',
                marginBottom: '4px',
              }}
            >
              {t('crm.rewards.pointsCost')}
            </label>
            <input
              type="number"
              min={1}
              value={form.pointsCost}
              onChange={(e) => setForm((f) => ({ ...f, pointsCost: Number(e.target.value) }))}
              style={inputStyle}
            />
          </div>
          <div style={{ flex: 1 }}>
            <label
              style={{
                fontSize: '11px',
                color: 'var(--hub-text-muted)',
                display: 'block',
                marginBottom: '4px',
              }}
            >
              {t('crm.rewards.stock')}
            </label>
            <input
              type="number"
              min={0}
              value={form.stock}
              onChange={(e) => setForm((f) => ({ ...f, stock: Number(e.target.value) }))}
              style={inputStyle}
            />
          </div>
        </div>
        <div style={{ display: 'flex', gap: '8px', justifyContent: 'flex-end' }}>
          <Button variant="outline" onClick={onClose}>
            {t('crmCommon.cancel')}
          </Button>
          <Button
            variant="primary"
            onClick={handleSubmit}
            disabled={upsert.isPending || !form.name.trim()}
          >
            {upsert.isPending ? 'Saving…' : 'Create reward'}
          </Button>
        </div>
      </div>
    </div>
  );
};

const LoyaltyRewards: React.FC = () => {
  const { t } = useTranslation();
  const canConfigureLoyalty = canCrmPermission('crm:loyalty:manage');
  const { data: fetched, isLoading, refetch } = useLoyaltyRewards();
  const rewards = fetched ?? [];
  const [showAdd, setShowAdd] = useState(false);

  return (
    <>
      <HelmetContainer title={`${t('crm.rewards.pageTitle')} · rtylr`} />
      <div style={{ maxWidth: '1100px' }}>
        <PageHeader
          title={t('crm.rewards.pageTitle')}
          subtitle={t('crm.rewards.subtitle')}
          actions={
            canConfigureLoyalty ? (
              <Button
                variant="primary"
                aria-label={t('crm.rewards.addReward')}
                onClick={() => setShowAdd(true)}
              >
                <Plus size={13} />
                {t('crm.rewards.addReward')}
              </Button>
            ) : undefined
          }
        />

        {showAdd && <AddRewardModal onClose={() => setShowAdd(false)} />}

        {isLoading ? (
          <Card>
            <SkeletonList count={4} />
          </Card>
        ) : rewards.length === 0 ? (
          <Card>
            <EmptyState
              icon={Gift}
              title={t('crm.rewards.empty')}
              description={t('crm.rewards.emptyDescription')}
              action={
                canConfigureLoyalty
                  ? { label: t('crm.rewards.addReward'), onClick: () => setShowAdd(true) }
                  : { label: t('crm.actions.refresh'), onClick: () => refetch() }
              }
            />
          </Card>
        ) : (
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fill, minmax(240px, 1fr))',
              gap: '14px',
            }}
          >
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
                  <div style={{ fontSize: '14px', fontWeight: 700, color: 'var(--hub-text)' }}>
                    {r.name}
                  </div>
                  <div
                    style={{ fontSize: '11px', color: 'var(--hub-text-muted)', marginTop: '2px' }}
                  >
                    {r.description}
                  </div>
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
                  <span
                    style={{
                      fontSize: '16px',
                      fontWeight: 700,
                      color: 'var(--crm-accent)',
                    }}
                  >
                    {r.pointsCost.toLocaleString()} pts
                  </span>
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
                    {r.stock === null ? 'Unlimited' : `${r.stock} left`}
                  </span>
                </div>
                <div
                  style={{
                    fontSize: '10px',
                    color: 'var(--hub-text-dim)',
                    textTransform: 'uppercase',
                    letterSpacing: '0.05em',
                  }}
                >
                  {r.category}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </>
  );
};

export default LoyaltyRewards;
