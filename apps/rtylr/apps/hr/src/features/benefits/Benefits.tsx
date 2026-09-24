import React from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { Shield, ClipboardList } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import {
  useBenefits,
  useMyBenefitEnrollments,
  useEnrollInBenefit,
} from '../../shared/hooks/useHrExtras';
import { Card, QueryError, SectionHeader, Skeleton } from '../../shared/components/ui';
import BenefitPlanCard from '../../shared/components/BenefitPlanCard';

const Benefits = (): React.JSX.Element => {
  const { t } = useTranslation();
  const {
    data: plans,
    isPending: plansPending,
    isError: plansError,
    isFetching: plansFetching,
    refetch: refetchPlans,
  } = useBenefits();
  const {
    data: enrollments = [],
    isPending: enrollmentsPending,
    isError: enrollmentsError,
    isFetching: enrollmentsFetching,
    refetch: refetchEnrollments,
  } = useMyBenefitEnrollments();
  const enroll = useEnrollInBenefit();
  const enrollInFlightRef = React.useRef(false);

  const loading = plansPending || enrollmentsPending;
  const queryError = plansError || enrollmentsError;
  const retry = (): void => {
    void refetchPlans();
    void refetchEnrollments();
  };

  const handleEnroll = (planId: string): void => {
    if (enrollInFlightRef.current || enroll.isPending) return;
    enrollInFlightRef.current = true;
    enroll.mutate(planId, {
      onSettled: () => {
        enrollInFlightRef.current = false;
      },
    });
  };

  return (
    <>
      <HelmetContainer title={pageTitle(t('nav.benefits'))} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <SectionHeader title={t('benefits.title')} subtitle={t('benefits.subtitle')} />

        {queryError ? (
          <QueryError onRetry={retry} isRetrying={plansFetching || enrollmentsFetching} />
        ) : loading ? (
          <Skeleton height={200} />
        ) : !plans || plans.length === 0 ? (
          <EmptyState
            icon={Shield}
            title={t('benefits.emptyTitle')}
            description={t('benefits.emptyDescription')}
            action={{ label: t('common.refresh'), onClick: () => refetchPlans() }}
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
              // A cancelled row still occupies the unique plan slot. Treating
              // every existing row as enrolled prevents a doomed re-enrollment.
              const enrolled = enrollments.some((e) => e.planId === String(p.id));
              return (
                <BenefitPlanCard
                  key={p.id}
                  plan={p}
                  enrolled={enrolled}
                  onEnroll={() => handleEnroll(String(p.id))}
                />
              );
            })}
          </div>
        )}

        {!queryError && (
          <Card>
            <SectionHeader title={t('benefits.activeEnrollments')} />
            <div style={{ marginTop: 10, fontSize: 12 }}>
              {enrollments.length === 0 && (
                <EmptyState
                  icon={ClipboardList}
                  title={t('benefits.noEnrollments')}
                  action={{ label: t('common.refresh'), onClick: () => refetchEnrollments() }}
                  size="compact"
                />
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
        )}
      </div>
    </>
  );
};

export default Benefits;
