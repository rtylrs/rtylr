import React, { useMemo } from 'react';
import { Eye } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import type { Task } from '../../shared/stores/flowStore';
import { useFlowStore } from '../../shared/stores/flowStore';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { useNavigate } from '@tanstack/react-router';
import { useTasks } from '../tasks/useTasks';
import TaskRow from '../../shared/components/TaskRow';
import { apiTaskToTask } from '../../shared/lib/taskMapper';

const Triage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { data: apiTasks } = useTasks({});
  const tasks: Task[] = (apiTasks ?? []).map(apiTaskToTask);
  const { selectedTaskIds } = useFlowStore();

  const needsTriage = useMemo(() => {
    return tasks.filter(
      (t) =>
        !t.archived &&
        t.status !== 'done' &&
        t.status !== 'cancelled' &&
        (t.priority === 'no_priority' || !t.dueDate || t.status === 'backlog')
    );
  }, [tasks]);

  return (
    <div className="flex flex-col h-full">
      <HelmetContainer title={`${t('flow.shell.triage')} · Flow`} />
      <div className="px-6 py-4 shrink-0" style={{ borderBottom: '1px solid var(--flow-border)' }}>
        <h1 className="text-title font-semibold" style={{ color: 'var(--flow-text)' }}>
          {t('flow.shell.triage')}
        </h1>
        <p className="text-meta" style={{ color: 'var(--flow-text-dim)' }}>
          {t('flow.triage.subtitle', { count: needsTriage.length })}
        </p>
      </div>
      <div className="flex-1 overflow-y-auto">
        {needsTriage.length === 0 ? (
          <EmptyState
            icon={Eye}
            title={t('flow.triage.emptyTitle')}
            description={t('flow.triage.emptyDescription')}
            action={{
              label: t('flow.triage.emptyCta'),
              onClick: () => navigate({ to: '/my-issues' }),
            }}
          />
        ) : (
          needsTriage.map((t) => (
            <TaskRow key={t.id} task={t} showProject selected={selectedTaskIds.includes(t.id)} />
          ))
        )}
      </div>
    </div>
  );
};

export default Triage;
