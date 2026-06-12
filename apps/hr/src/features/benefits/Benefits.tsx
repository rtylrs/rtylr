import React from 'react';
import { toast } from 'sonner';
import { HelmetContainer } from '@voxire/components';
import { Shield } from 'lucide-react';
import {
  useBenefits,
  useBenefitEnrollments,
  useEnrollBenefit,
} from '../../shared/hooks/useHrExtras';
import { Card, SectionHeader, EmptyState, Skeleton } from '../../shared/components/ui';
import BenefitPlanCard from '../../shared/components/BenefitPlanCard';

const Benefits: React.FC = () => {
  const { data: plans, isLoading } = useBenefits();
  const { data: enrollments = [] } = useBenefitEnrollments();
  const enroll = useEnrollBenefit();

  return (
    <>
      <HelmetContainer title="Benefits · rtylr" />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <SectionHeader title="Benefits" subtitle="Health, dental, vision, 401k and more" />
        {isLoading ? (
          <Skeleton height={200} />
        ) : !plans || plans.length === 0 ? (
          <EmptyState
            icon={<Shield size={20} />}
            title="No benefit plans"
            description="Create plans to allow enrollment."
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
              const enrolled = enrollments.some((e) => e.planId === p.id && e.staffId === 1);
              return (
                <BenefitPlanCard
                  key={p.id}
                  plan={p}
                  enrolled={enrolled}
                  onEnroll={() =>
                    enroll.mutate(
                      { staffId: 1, staffName: 'Alex Chen', planId: p.id, planName: p.name },
                      { onSuccess: () => toast.success('Enrolled') }
                    )
                  }
                />
              );
            })}
          </div>
        )}

        <Card>
          <SectionHeader title="Active enrollments" />
          <div style={{ marginTop: 10, fontSize: 12 }}>
            {enrollments.length === 0 && (
              <p style={{ color: 'var(--hr-text-muted)' }}>No enrollments yet.</p>
            )}
            {enrollments.map((e) => (
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
                  {e.staffName} - {e.planName}
                </span>
                <span style={{ color: 'var(--hr-text-muted)' }}>
                  {e.status} since {e.startedAt}
                </span>
              </div>
            ))}
          </div>
        </Card>
      </div>
    </>
  );
};

export default Benefits;
