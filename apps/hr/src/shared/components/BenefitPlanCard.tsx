import React from 'react';
import { useTranslation } from 'react-i18next';
import { Shield, Check } from 'lucide-react';
import { Card, ACCENT } from './ui';
import { BenefitPlan } from '../schemas/benefits';

// category is the backend's free-form BenefitPlan.type, title-cased - not a
// fixed client enum - so this is a display default per known value, not an
// exhaustive map. DEFAULT_COLOR covers any category a business configures
// that isn't one of these common ones.
const CATEGORY_COLORS: Record<string, string> = {
  Health: '#ef4444',
  Dental: '#3b82f6',
  Vision: '#10b981',
  '401k': '#f59e0b',
  Life: '#8b5cf6',
};
const DEFAULT_CATEGORY_COLOR = '#6b7280';

interface Props {
  plan: BenefitPlan;
  enrolled?: boolean;
  onEnroll?: () => void;
}

const BenefitPlanCard: React.FC<Props> = ({ plan, enrolled, onEnroll }) => {
  const { t } = useTranslation();
  const color = CATEGORY_COLORS[plan.category] ?? DEFAULT_CATEGORY_COLOR;
  return (
    <Card style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
        <span style={{ padding: 8, borderRadius: 10, background: `${color}22`, color }}>
          <Shield size={16} />
        </span>
        <div style={{ flex: 1 }}>
          <div style={{ fontSize: 13, fontWeight: 700, color: 'var(--hr-text)' }}>{plan.name}</div>
          <div style={{ fontSize: 11, color: 'var(--hr-text-muted)' }}>
            {plan.carrier} • {plan.category}
          </div>
        </div>
      </div>
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          paddingTop: 8,
          borderTop: '1px solid var(--hr-border)',
        }}
      >
        <div>
          <div style={{ fontSize: 11, color: 'var(--hr-text-muted)' }}>{t('benefits.monthly')}</div>
          <div style={{ fontSize: 15, fontWeight: 700, color: 'var(--hr-text)' }}>
            ${plan.monthlyCost}
          </div>
        </div>
        {enrolled ? (
          <span
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: 4,
              padding: '4px 10px',
              borderRadius: 8,
              background: 'rgba(16,185,129,0.15)',
              color: '#10b981',
              fontSize: 11,
              fontWeight: 600,
            }}
          >
            <Check size={12} /> {t('benefits.enrolled')}
          </span>
        ) : (
          <button
            type="button"
            onClick={onEnroll}
            style={{
              padding: '6px 12px',
              borderRadius: 8,
              border: `1px solid ${ACCENT}55`,
              background: `${ACCENT}22`,
              color: ACCENT,
              fontSize: 11,
              fontWeight: 600,
              cursor: 'pointer',
            }}
          >
            {t('benefits.enroll')}
          </button>
        )}
      </div>
    </Card>
  );
};

export default BenefitPlanCard;
