import { EmptyState } from '../../../../shared/components/dash/EmptyState';
import React, { useMemo } from 'react';
import { Eye } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import type { Task } from '../../shared/stores/flowStore';
import { useFlowStore } from '../../shared/stores/flowStore';
import { HelmetContainer } from '@voxire/components';

import { useNavigate } from '@tanstack/react-router';
import { useTasks } from '../tasks/useTasks';
import TaskRow from '../../shared/components/TaskRow';
import { apiTaskToTask } from '../../shared/lib/taskMapper';
import { hasQueryFailed } from '../../shared/lib/queryFailureUtils';
import { ListLoadFailed } from '../../shared/components/ListLoadFailed';
import { TaskRowSkeleton } from '../../shared/components/Skeleton';

const Triage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const tasksQuery = useTasks({});
  const { data: apiTasks, isLoading } = tasksQuery;
  const loadFailed = hasQueryFailed(tasksQuery);
  // `apiTasks` is undefined until the query resolves, which made `needsTriage`
  // empty and the screen say "Inbox zero" and "0 tasks need attention" before
  // anything had been fetched. A confident sentence should only come from a
  // query that ran, so the count and the empty state both wait for it.
  const resolved = !isLoading && !loadFailed;
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
        <h1 className="text-flow-title font-semibold" style={{ color: 'var(--flow-text)' }}>
          {t('flow.shell.triage')}
        </h1>
        {resolved && (
          <p className="text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
            {t('flow.triage.subtitle', { count: needsTriage.length })}
          </p>
        )}
        {/* 13px and plain words. This was 12px, the smallest text on the
            screen, and used an arrow notation used nowhere else in Flow. */}
        <p className="text-flow-body mt-1" style={{ color: 'var(--flow-text-muted)' }}>
          {t('flow.triage.explanation')}
        </p>
      </div>
      <div className="flex-1 overflow-y-auto">
        {isLoading ? (
          <div className="px-6 py-4">
            <TaskRowSkeleton />
            <TaskRowSkeleton />
            <TaskRowSkeleton />
          </div>
        ) : loadFailed ? (
          <ListLoadFailed onRetry={() => void tasksQuery.refetch()} />
        ) : needsTriage.length === 0 ? (
          <EmptyState
            icon={Eye}
            title={t('flow.triage.emptyTitle')}
            description={t('flow.triage.emptyDescription')}
            action={{
              label: t('flow.triage.emptyCta'),
              onClick: () => navigate({ to: '/flow/my-issues' }),
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
