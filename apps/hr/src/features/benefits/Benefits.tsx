import React from 'react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { Shield } from 'lucide-react';
import {
  useBenefits,
  useMyBenefitEnrollments,
  useEnrollInBenefit,
} from '../../shared/hooks/useHrExtras';
import { Card, SectionHeader, EmptyState, Skeleton } from '../../shared/components/ui';
import BenefitPlanCard from '../../shared/components/BenefitPlanCard';

const Benefits = (): React.JSX.Element => {
  const { t } = useTranslation();
  const { data: plans, isLoading: plansLoading } = useBenefits();
  const { data: enrollments = [], isLoading: enrollmentsLoading } = useMyBenefitEnrollments();
  const enroll = useEnrollInBenefit();

  const loading = plansLoading || enrollmentsLoading;

  return (
    <>
      <HelmetContainer title="Benefits · rtylr" />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <SectionHeader title={t('benefits.title')} subtitle={t('benefits.subtitle')} />

        {loading ? (
          <Skeleton height={200} />
        ) : !plans || plans.length === 0 ? (
          <EmptyState
            icon={<Shield size={20} />}
            title={t('benefits.emptyTitle')}
            description={t('benefits.emptyDescription')}
          />
        ) : (
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))',
              gap: 12,
            }}
          >
            {plans.map((p) => {
              // Enrollments are the caller's own, resolved server-side - a
              // cancelled enrollment for this plan should not read as active.
              const enrolled = enrollments.some(
                (e) => e.planId === String(p.id) && e.status === 'active'
              );
              return (
                <BenefitPlanCard
                  key={p.id}
                  plan={p}
                  enrolled={enrolled}
                  onEnroll={() => enroll.mutate(String(p.id))}
                />
              );
            })}
          </div>
        )}

        <Card>
          <SectionHeader title={t('benefits.activeEnrollments')} />
          <div style={{ marginTop: 10, fontSize: 12 }}>
            {enrollments.length === 0 && (
              <p style={{ color: 'var(--hr-text-muted)' }}>{t('benefits.noEnrollments')}</p>
            )}
            {enrollments.map((e) => {
              const plan = plans?.find((p) => String(p.id) === e.planId);
              return (
                <div
                  key={e.id}
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    padding: '8px 0',
                    borderBottom: '1px solid var(--hr-border)',
                  }}
                >
                  <span style={{ color: 'var(--hr-text)' }}>
                    {plan?.name ?? t('benefits.unknownPlan')}
                  </span>
                  <span style={{ color: 'var(--hr-text-muted)' }}>
                    {t('benefits.statusSince', { status: e.status, date: e.enrolledAt })}
                  </span>
                </div>
              );
            })}
          </div>
        </Card>
      </div>
    </>
  );
};

export default Benefits;
