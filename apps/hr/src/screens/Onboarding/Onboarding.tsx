import React from 'react';
import { toast } from 'sonner';
import { Rocket } from 'lucide-react';
import { useOnboardingChecklist, useToggleOnboardingTask } from '../../hooks/useHrExtras';
import { Card, SectionHeader, EmptyState, Skeleton, Pill, Checkbox, ACCENT } from '../../components/ui';

const CATEGORIES: Record<string, string> = {
  Paperwork: '#3b82f6',
  Access: '#8b5cf6',
  Training: '#10b981',
  'Meet the team': '#f59e0b',
  Equipment: '#ef4444',
};

const Onboarding: React.FC = () => {
  const { data, isLoading } = useOnboardingChecklist();
  const toggle = useToggleOnboardingTask();

  const grouped = (data ?? []).reduce<Record<string, typeof data>>((acc, t) => {
    acc[t.staffName] = acc[t.staffName] || ([] as any);
    (acc[t.staffName] as any).push(t);
    return acc;
  }, {});

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <SectionHeader title="Onboarding" subtitle="New hire workflow checklist" />
      {isLoading ? (
        <Skeleton height={240} />
      ) : !data || data.length === 0 ? (
        <EmptyState icon={<Rocket size={20} />} title="No onboarding in progress" description="When a new hire starts, their tasks will show up here." />
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          {Object.entries(grouped).map(([name, tasks]) => {
            const list = tasks as typeof data;
            const done = list!.filter((t) => t.completed).length;
            const pct = Math.round((done / list!.length) * 100);
            return (
              <Card key={name}>
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 10 }}>
                  <div>
                    <div style={{ fontSize: 14, fontWeight: 700, color: 'var(--hr-text)' }}>{name}</div>
                    <div style={{ fontSize: 11, color: 'var(--hr-text-muted)' }}>{done} of {list!.length} tasks done</div>
                  </div>
                  <div style={{ fontSize: 18, fontWeight: 700, color: ACCENT }}>{pct}%</div>
                </div>
                <div style={{ height: 4, borderRadius: 999, background: 'var(--hr-hover)', marginBottom: 12 }}>
                  <div style={{ width: `${pct}%`, height: '100%', background: ACCENT, borderRadius: 999 }} />
                </div>
                <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
                  {list!.map((t) => (
                    <div key={t.id} style={{ display: 'flex', alignItems: 'center', gap: 10, padding: 8, borderRadius: 8, background: 'var(--hr-hover)' }}>
                      <Checkbox
                        checked={t.completed}
                        onChange={() => toggle.mutate(t.id, { onSuccess: () => toast.success(t.completed ? 'Marked incomplete' : 'Marked complete') })}
                      />
                      <div style={{ flex: 1, fontSize: 12, color: 'var(--hr-text)', textDecoration: t.completed ? 'line-through' : 'none', opacity: t.completed ? 0.6 : 1 }}>
                        {t.task}
                      </div>
                      <Pill color={CATEGORIES[t.category] ?? '#a1a1aa'}>{t.category}</Pill>
                      {t.dueDate && <span style={{ fontSize: 10, color: 'var(--hr-text-muted)' }}>Due {t.dueDate}</span>}
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
