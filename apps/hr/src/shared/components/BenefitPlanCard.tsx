import React from 'react';
import { useTranslation } from 'react-i18next';
import { Shield, Check } from 'lucide-react';
import { Card, ACCENT } from './ui';
import { BenefitPlan } from '../services/hrExtras';

const CATEGORY_COLORS: Record<BenefitPlan['category'], string> = {
  Health: '#ef4444',
  Dental: '#3b82f6',
  Vision: '#10b981',
  '401k': '#f59e0b',
  Life: '#8b5cf6',
};

interface Props {
  plan: BenefitPlan;
  enrolled?: boolean;
  onEnroll?: () => void;
}

const BenefitPlanCard: React.FC<Props> = ({ plan, enrolled, onEnroll }) => {
  const { t } = useTranslation();
  const color = CATEGORY_COLORS[plan.category];
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
        <div>
          <div style={{ fontSize: 11, color: 'var(--hr-text-muted)' }}>
            {t('benefits.enrolled')}
          </div>
          <div style={{ fontSize: 13, fontWeight: 600, color: 'var(--hr-text)' }}>
            {plan.enrolledCount}
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
