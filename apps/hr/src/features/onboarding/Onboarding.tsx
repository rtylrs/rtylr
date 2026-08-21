import React from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { Rocket } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { useOnboardingChecklist, useToggleOnboardingTask } from '../../shared/hooks/useHrExtras';
import { Card, SectionHeader, Skeleton, Pill, Checkbox, ACCENT } from '../../shared/components/ui';

const CATEGORIES: Record<string, string> = {
  Paperwork: 'var(--hr-shift-blue)',
  Access: 'var(--hr-accent)',
  Training: 'var(--hr-onboarding-complete)',
  'Meet the team': 'var(--hr-status-pending)',
  Equipment: 'var(--hr-error)',
};

const pillStyle = (cssVar: string): React.CSSProperties => ({
  background: `color-mix(in srgb, ${cssVar} 13%, transparent)`,
  color: cssVar,
  border: `1px solid color-mix(in srgb, ${cssVar} 20%, transparent)`,
});

const Onboarding: React.FC = () => {
  const { t } = useTranslation();
  const { data, isLoading, refetch } = useOnboardingChecklist();
  const toggle = useToggleOnboardingTask();

  const grouped = (data ?? []).reduce<Record<string, typeof data>>((acc, task) => {
    const key = task.staffName;
    acc[key] = [...(acc[key] ?? []), task];
    return acc;
  }, {});

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <SectionHeader title={t('onboarding.title')} subtitle={t('onboarding.subtitle')} />
      {isLoading ? (
        <Skeleton height={240} />
      ) : !data || data.length === 0 ? (
        <EmptyState
          icon={Rocket}
          title={t('onboarding.emptyTitle')}
          description={t('onboarding.emptyDescription')}
          action={{ label: t('common.refresh'), onClick: () => refetch() }}
        />
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          {Object.entries(grouped).map(([name, tasks]) => {
            const list = tasks as typeof data;
            const done = list!.filter((task) => task.completed).length;
            const pct = Math.round((done / list!.length) * 100);
            return (
              <Card key={name}>
                <div
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    marginBottom: 10,
                  }}
                >
                  <div>
                    <div style={{ fontSize: 14, fontWeight: 700, color: 'var(--hr-text)' }}>
                      {name}
                    </div>
                    <div style={{ fontSize: 11, color: 'var(--hr-text-muted)' }}>
                      {t('onboarding.tasksDone', { done, total: list!.length })}
                    </div>
                  </div>
                  <div style={{ fontSize: 18, fontWeight: 700, color: ACCENT }}>{pct}%</div>
                </div>
                <div
                  style={{
                    height: 4,
                    borderRadius: 999,
                    background: 'var(--hr-hover)',
                    marginBottom: 12,
                  }}
                >
                  <div
                    style={{
                      width: `${pct}%`,
                      height: '100%',
                      background: ACCENT,
                      borderRadius: 999,
                    }}
                  />
                </div>
                <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
                  {list!.map((task) => (
                    <div
                      key={task.id}
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: 10,
                        padding: 8,
                        borderRadius: 8,
                        background: 'var(--hr-hover)',
                      }}
                    >
                      <Checkbox
                        checked={task.completed}
                        onChange={() =>
                          toggle.mutate(task.id, {
                            onSuccess: () =>
                              toast.success(
                                task.completed
                                  ? t('onboarding.markedIncomplete')
                                  : t('onboarding.markedComplete')
                              ),
                          })
                        }
                      />
                      <div
                        style={{
                          flex: 1,
                          fontSize: 12,
                          color: 'var(--hr-text)',
                          textDecoration: task.completed ? 'line-through' : 'none',
                          opacity: task.completed ? 0.6 : 1,
                        }}
                      >
                        {task.task}
                      </div>
                      <Pill
                        style={pillStyle(CATEGORIES[task.category] ?? 'var(--hr-status-draft)')}
                      >
                        {task.category}
                      </Pill>
                      {task.dueDate && (
                        <span style={{ fontSize: 10, color: 'var(--hr-text-muted)' }}>
                          {t('onboarding.due', { date: task.dueDate })}
                        </span>
                      )}
                    </div>
                  ))}
                </div>
              </Card>
            );
          })}
        </div>
      )}
    </div>
  );
};

export default Onboarding;
