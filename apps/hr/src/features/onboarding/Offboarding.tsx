import React from 'react';
import { toast } from 'sonner';
import { LogOut } from 'lucide-react';
import { useOffboardingChecklist, useToggleOffboardingTask } from '../../shared/hooks/useHrExtras';
import {
  Card,
  SectionHeader,
  EmptyState,
  Skeleton,
  Pill,
  Checkbox,
} from '../../shared/components/ui';

const CATEGORIES: Record<string, string> = {
  'Knowledge transfer': '#3b82f6',
  Access: '#ef4444',
  Equipment: '#f59e0b',
  Paperwork: '#8b5cf6',
};

const Offboarding: React.FC = () => {
  const { data, isLoading } = useOffboardingChecklist();
  const toggle = useToggleOffboardingTask();

  const grouped = (data ?? []).reduce<Record<string, typeof data>>((acc, t) => {
    acc[t.staffName] = acc[t.staffName] || ([] as any);
    (acc[t.staffName] as any).push(t);
    return acc;
  }, {});

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <SectionHeader title="Offboarding" subtitle="Exit checklist for departing employees" />
      {isLoading ? (
        <Skeleton height={200} />
      ) : !data || data.length === 0 ? (
        <EmptyState
          icon={<LogOut size={20} />}
          title="No offboarding in progress"
          description="Departing employees' checklists will show here."
        />
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          {Object.entries(grouped).map(([name, tasks]) => {
            const list = tasks as typeof data;
            const done = list!.filter((t) => t.completed).length;
            return (
              <Card key={name}>
                <div style={{ marginBottom: 10 }}>
                  <div style={{ fontSize: 14, fontWeight: 700, color: 'var(--hr-text)' }}>
                    {name}
                  </div>
                  <div style={{ fontSize: 11, color: 'var(--hr-text-muted)' }}>
                    {done} of {list!.length} complete
                  </div>
                </div>
                <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
                  {list!.map((t) => (
                    <div
                      key={t.id}
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
                        checked={t.completed}
                        onChange={() =>
                          toggle.mutate(t.id, { onSuccess: () => toast.success('Updated') })
                        }
                      />
                      <div
                        style={{
                          flex: 1,
                          fontSize: 12,
                          color: 'var(--hr-text)',
                          textDecoration: t.completed ? 'line-through' : 'none',
                          opacity: t.completed ? 0.6 : 1,
                        }}
                      >
                        {t.task}
                      </div>
                      <Pill color={CATEGORIES[t.category] ?? '#a1a1aa'}>{t.category}</Pill>
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

export default Offboarding;
