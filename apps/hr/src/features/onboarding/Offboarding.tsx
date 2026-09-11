import React from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { HelmetContainer } from '@voxire/components';
import { LogOut } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { useOffboardingChecklist, useToggleOffboardingTask } from '../../shared/hooks/useHrExtras';
import { Card, SectionHeader, Skeleton, Pill, Checkbox } from '../../shared/components/ui';

const CATEGORIES: Record<string, string> = {
  'Knowledge transfer': 'var(--hr-shift-blue)',
  Access: 'var(--hr-error)',
  Equipment: 'var(--hr-status-pending)',
  Paperwork: 'var(--hr-accent)',
};

const pillStyle = (cssVar: string): React.CSSProperties => ({
  background: `color-mix(in srgb, ${cssVar} 13%, transparent)`,
  color: cssVar,
  border: `1px solid color-mix(in srgb, ${cssVar} 20%, transparent)`,
});

const Offboarding: React.FC = () => {
  const { t } = useTranslation();
  const { data, isLoading, refetch } = useOffboardingChecklist();
  const toggle = useToggleOffboardingTask();

  const grouped = (data ?? []).reduce<Record<string, typeof data>>((acc, task) => {
    const key = task.staffName;
    acc[key] = [...(acc[key] ?? []), task];
    return acc;
  }, {});

  return (
    <>
      <HelmetContainer title={pageTitle(t('nav.offboarding'))} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <SectionHeader title={t('offboarding.title')} subtitle={t('offboarding.subtitle')} />
        {isLoading ? (
          <Skeleton height={200} />
        ) : !data || data.length === 0 ? (
          <EmptyState
            icon={LogOut}
            title={t('offboarding.emptyTitle')}
            description={t('offboarding.emptyDescription')}
            action={{ label: t('common.refresh'), onClick: () => refetch() }}
          />
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            {Object.entries(grouped).map(([name, tasks]) => {
              const list = tasks as typeof data;
              const done = list!.filter((task) => task.completed).length;
              return (
                <Card key={name}>
                  <div style={{ marginBottom: 10 }}>
                    <div style={{ fontSize: 14, fontWeight: 700, color: 'var(--hr-text)' }}>
                      {name}
                    </div>
                    <div style={{ fontSize: 11, color: 'var(--hr-text-muted)' }}>
                      {t('offboarding.tasksComplete', { done, total: list!.length })}
                    </div>
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
                              onSuccess: () => toast.success(t('offboarding.updated')),
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
                      </div>
                    ))}
                  </div>
                </Card>
              );
            })}
          </div>
        )}
      </div>
    </>
  );
};

export default Offboarding;
