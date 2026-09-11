import React from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { GraduationCap, Clock, Layers, Award } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { useTraining, useTrainingProgress } from '../../shared/hooks/useHrExtras';
import { Card, SectionHeader, Skeleton, Pill, ACCENT } from '../../shared/components/ui';
import TrainingProgress from '../../shared/components/TrainingProgress';

const CATEGORY_COLORS: Record<string, string> = {
  Compliance: '#ef4444',
  Safety: '#f59e0b',
  Skills: '#3b82f6',
  Leadership: '#8b5cf6',
};

const Training: React.FC = () => {
  const { t } = useTranslation();
  const { data: courses, isLoading, refetch } = useTraining();
  const { data: progress = [] } = useTrainingProgress();

  return (
    <>
      <HelmetContainer title={pageTitle(t('nav.training'))} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <SectionHeader title="Training" subtitle="Courses, completion tracking, certificates" />

        {isLoading ? (
          <Skeleton height={240} />
        ) : !courses || courses.length === 0 ? (
          <EmptyState
            icon={GraduationCap}
            title={t('training.emptyTitle')}
            description={t('training.emptyDescription')}
            action={{ label: t('common.refresh'), onClick: () => refetch() }}
          />
        ) : (
          <>
            <div
              style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(auto-fill, minmax(260px, 1fr))',
                gap: 12,
              }}
            >
              {courses.map((c) => (
                <Card key={c.id}>
                  <div style={{ display: 'flex', alignItems: 'start', gap: 10 }}>
                    <span
                      style={{
                        padding: 8,
                        borderRadius: 10,
                        background: `${CATEGORY_COLORS[c.category] ?? ACCENT}22`,
                        color: CATEGORY_COLORS[c.category] ?? ACCENT,
                      }}
                    >
                      <GraduationCap size={16} />
                    </span>
                    <div style={{ flex: 1 }}>
                      <div style={{ fontSize: 13, fontWeight: 700, color: 'var(--hr-text)' }}>
                        {c.title}
                      </div>
                      <div style={{ display: 'flex', gap: 4, marginTop: 4 }}>
                        <Pill color={CATEGORY_COLORS[c.category] ?? ACCENT}>{c.category}</Pill>
                        {c.required && <Pill color="#ef4444">Required</Pill>}
                      </div>
                    </div>
                  </div>
                  <div
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 12,
                      marginTop: 12,
                      fontSize: 11,
                      color: 'var(--hr-text-muted)',
                    }}
                  >
                    <span style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
                      <Clock size={11} /> {c.durationMin} min
                    </span>
                    <span style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
                      <Layers size={11} /> {c.modules} modules
                    </span>
                  </div>
                </Card>
              ))}
            </div>

            <Card>
              <SectionHeader title="Employee progress" />
              <div style={{ display: 'flex', flexDirection: 'column', gap: 12, marginTop: 12 }}>
                {progress.length === 0 ? (
                  <p style={{ fontSize: 11, color: 'var(--hr-text-muted)' }}>
                    {t('training.noEnrollments')}
                  </p>
                ) : (
                  progress.map((p) => (
                    <div key={p.id} style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                      <div
                        style={{
                          minWidth: 120,
                          fontSize: 12,
                          color: 'var(--hr-text)',
                          fontWeight: 600,
                        }}
                      >
                        {p.staffName}
                      </div>
                      <div style={{ flex: 1 }}>
                        <TrainingProgress progress={p.progress} label={p.courseName} />
                      </div>
                      {p.completedAt && <Award size={14} style={{ color: '#10b981' }} />}
                    </div>
                  ))
                )}
              </div>
            </Card>
          </>
        )}
      </div>
    </>
  );
};

export default Training;
