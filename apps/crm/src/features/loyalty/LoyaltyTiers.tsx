import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { toast } from 'sonner';
import { useUserStore } from '@voxire/auth';
import { Award, Check, Pencil, X } from 'lucide-react';
import { useLoyaltyTiers, useUpsertTier } from '../../shared/hooks/queries/useLoyalty';
import { type LoyaltyTier } from '../../shared/schemas/loyalty';
import { PageHeader, Card, SkeletonList } from '../../shared/components/ui';
import { canCrmPermission } from '../../shared/lib/crmPermissions';

interface TierEditFormProps {
  tier: LoyaltyTier;
  allTiers: LoyaltyTier[];
  onClose: () => void;
}

const TierEditForm: React.FC<TierEditFormProps> = ({ tier, allTiers, onClose }) => {
  const { t } = useTranslation();
  const upsert = useUpsertTier();
  const [threshold, setThreshold] = useState(tier.threshold);
  const [multiplier, setMultiplier] = useState(tier.multiplier);

  const handleSave = (): void => {
    const sorted = [...allTiers].sort((a, b) => a.threshold - b.threshold);
    const currentIndex = sorted.findIndex((t) => t.id === tier.id);
    const lowerTier = currentIndex > 0 ? sorted[currentIndex - 1] : null;
    const higherTier = currentIndex < sorted.length - 1 ? sorted[currentIndex + 1] : null;

    if (lowerTier !== null && threshold <= lowerTier.threshold) {
      toast.error(t('crm.loyaltyTiers.thresholdOrder'));
      return;
    }
    if (higherTier !== null && threshold >= higherTier.threshold) {
      toast.error(t('crm.loyaltyTiers.thresholdOrder'));
      return;
    }

    upsert.mutate({ id: tier.id, data: { threshold, multiplier } }, { onSuccess: onClose });
  };

  return (
    <div
      style={{
        marginTop: '14px',
        padding: '12px',
        background: 'var(--hub-hover)',
        borderRadius: '10px',
        display: 'flex',
        flexDirection: 'column',
        gap: '10px',
      }}
    >
      <div>
        <label
          style={{
            display: 'block',
            fontSize: '10px',
            fontWeight: 600,
            color: 'var(--hub-text-dim)',
            textTransform: 'uppercase',
            letterSpacing: '0.05em',
            marginBottom: '4px',
          }}
        >
          Threshold (points)
        </label>
        <input
          type="number"
          value={threshold}
          min={0}
          onChange={(e) => setThreshold(Number(e.target.value))}
          style={{
            width: '100%',
            height: '32px',
            padding: '0 8px',
            borderRadius: '6px',
            border: '1px solid var(--hub-border)',
            background: 'var(--hub-input-bg)',
            color: 'var(--hub-text)',
            fontSize: '13px',
            outline: 'none',
            boxSizing: 'border-box',
          }}
        />
      </div>
      <div>
        <label
          style={{
            display: 'block',
            fontSize: '10px',
            fontWeight: 600,
            color: 'var(--hub-text-dim)',
            textTransform: 'uppercase',
            letterSpacing: '0.05em',
            marginBottom: '4px',
          }}
        >
          Discount (%)
        </label>
        <input
          type="number"
          value={multiplier}
          min={0}
          max={100}
          step={0.1}
          onChange={(e) => setMultiplier(Number(e.target.value))}
          style={{
            width: '100%',
            height: '32px',
            padding: '0 8px',
            borderRadius: '6px',
            border: '1px solid var(--hub-border)',
            background: 'var(--hub-input-bg)',
            color: 'var(--hub-text)',
            fontSize: '13px',
            outline: 'none',
            boxSizing: 'border-box',
          }}
        />
      </div>
      <div style={{ display: 'flex', gap: '6px', justifyContent: 'flex-end' }}>
        <button
          type="button"
          onClick={onClose}
          style={{
            padding: '5px 12px',
            borderRadius: '6px',
            fontSize: '12px',
            fontWeight: 500,
            background: 'var(--hub-card)',
            color: 'var(--hub-text-muted)',
            border: '1px solid var(--hub-border)',
            cursor: 'pointer',
          }}
        >
          {t('common.cancel')}
        </button>
        <button
          type="button"
          onClick={handleSave}
          disabled={upsert.isPending}
          style={{
            padding: '5px 12px',
            borderRadius: '6px',
            fontSize: '12px',
            fontWeight: 600,
            background: 'var(--accent)',
            color: 'var(--crm-on-accent)',
            border: 'none',
            cursor: upsert.isPending ? 'not-allowed' : 'pointer',
            opacity: upsert.isPending ? 0.7 : 1,
          }}
        >
          {upsert.isPending ? 'Saving…' : 'Save'}
        </button>
      </div>
    </div>
  );
};

const LoyaltyTiers: React.FC = () => {
  const { t } = useTranslation();
  const role = useUserStore((state) => state.user?.role ?? state.role);
  const canConfigureLoyalty = canCrmPermission(role, 'crm:loyalty:configure');
  const { data: fetched, isLoading } = useLoyaltyTiers();
  const tiers = (fetched ?? []).sort((a, b) => a.threshold - b.threshold);
  const [editingId, setEditingId] = useState<number | null>(null);

  return (
    <>
      <HelmetContainer title={`${t('crm.loyaltyTiers.pageTitle')} · rtylr`} />
      <div style={{ maxWidth: '1100px' }}>
        <PageHeader
          title={t('crm.loyaltyTiers.pageTitle')}
          subtitle={t('crm.loyaltyTiers.subtitle')}
        />

        {isLoading ? (
          <Card>
            <SkeletonList count={4} />
          </Card>
        ) : (
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fill, minmax(240px, 1fr))',
              gap: '14px',
            }}
          >
            {tiers.map((t) => (
              <div
                key={t.id}
                style={{
                  background: 'var(--hub-card)',
                  border: `1px solid ${t.color}40`,
                  borderRadius: '14px',
                  padding: '18px',
                  position: 'relative',
                  overflow: 'hidden',
                }}
              >
                <div
                  style={{
                    position: 'absolute',
                    top: 0,
                    left: 0,
                    right: 0,
                    height: '4px',
                    background: t.color,
                  }}
                />
                <div
                  style={{
                    display: 'flex',
                    alignItems: 'flex-start',
                    justifyContent: 'space-between',
                    marginBottom: '10px',
                  }}
                >
                  <div
                    style={{
                      width: '48px',
                      height: '48px',
                      borderRadius: '12px',
                      background: `${t.color}14`,
                      color: t.color,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                    }}
                  >
                    <Award size={22} />
                  </div>
                  {canConfigureLoyalty && (
                    <button
                      type="button"
                      aria-label={`Edit ${t.name} tier`}
                      onClick={() => setEditingId(editingId === t.id ? null : t.id)}
                      style={{
                        width: '28px',
                        height: '28px',
                        borderRadius: '6px',
                        border: '1px solid var(--hub-border)',
                        background: 'var(--hub-hover)',
                        color: 'var(--hub-text-muted)',
                        cursor: 'pointer',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        flexShrink: 0,
                      }}
                    >
                      {editingId === t.id ? <X size={13} /> : <Pencil size={13} />}
                    </button>
                  )}
                </div>
                <h3
                  style={{
                    fontSize: '18px',
                    fontWeight: 700,
                    color: 'var(--hub-text)',
                    margin: '0 0 4px',
                  }}
                >
                  {t.name}
                </h3>
                <div
                  style={{
                    fontSize: '11px',
                    color: 'var(--hub-text-dim)',
                    textTransform: 'uppercase',
                    letterSpacing: '0.05em',
                    fontWeight: 600,
                  }}
                >
                  {t.threshold === 0 ? 'Starting tier' : `${t.threshold}+ points`}
                </div>
                <div
                  style={{
                    fontSize: '12px',
                    color: 'var(--hub-text)',
                    marginTop: '10px',
                    padding: '6px 10px',
                    background: 'var(--hub-hover)',
                    borderRadius: '8px',
                    display: 'inline-block',
                  }}
                >
                  {t.multiplier}% discount
                </div>
                <div
                  style={{
                    marginTop: '14px',
                    display: 'flex',
                    flexDirection: 'column',
                    gap: '6px',
                  }}
                >
                  {t.perks.map((p) => (
                    <div
                      key={p}
                      style={{
                        display: 'flex',
                        alignItems: 'flex-start',
                        gap: '6px',
                        fontSize: '12px',
                        color: 'var(--hub-text-muted)',
                      }}
                    >
                      <Check
                        size={12}
                        style={{ color: t.color, flexShrink: 0, marginTop: '2px' }}
                      />
                      {p}
                    </div>
                  ))}
                </div>
                {editingId === t.id && (
                  <TierEditForm tier={t} allTiers={tiers} onClose={() => setEditingId(null)} />
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </>
  );
};

export default LoyaltyTiers;
