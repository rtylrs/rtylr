import React, { useMemo } from 'react';
import { useOpenTask } from '../../shared/lib/useOpenTask';
import { useNavigate } from '@tanstack/react-router';
import { Archive as ArchiveIcon, RotateCcw } from 'lucide-react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { useTasks, useUpdateTask } from '../tasks/useTasks';
import type { ApiTask } from '../../shared/services/flowApi';
import { archivedTasks } from '../../shared/lib/taskArchiveUtils';
import PriorityIcon from '../../shared/components/PriorityIcon';
import { hasQueryFailed } from '../../shared/lib/queryFailureUtils';
import { ListLoadFailed } from '../../shared/components/ListLoadFailed';
import { TaskRowSkeleton } from '../../shared/components/Skeleton';

const Archive: React.FC = () => {
  const { t } = useTranslation();
  const openTask = useOpenTask();
  const navigate = useNavigate();
  const tasksQuery = useTasks({ activeOnly: false });
  const { data: apiTasks = [], isLoading } = tasksQuery;
  const loadFailed = hasQueryFailed(tasksQuery);
  // The count and the empty state both assert a fact. apiTasks is [] until the
  // query resolves, so before this the screen said "0 archived tasks" and
  // "Nothing archived" while archived tasks existed.
  const resolved = !isLoading && !loadFailed;
  const restoreTask = useUpdateTask();

  const archived = useMemo(() => archivedTasks(apiTasks as ApiTask[]), [apiTasks]);

  return (
    <div className="flex flex-col h-full">
      <HelmetContainer title={`${t('archive.title')} · Flow`} />
      <div className="px-6 py-4 shrink-0" style={{ borderBottom: '1px solid var(--flow-border)' }}>
        <h1 className="text-title font-semibold" style={{ color: 'var(--flow-text)' }}>
          {t('archive.title')}
        </h1>
        {resolved && (
          <p className="text-meta" style={{ color: 'var(--flow-text-dim)' }}>
            {t('archive.count', { count: archived.length })}
          </p>
        )}
        {/* Archiving is not deleting, and nothing on this screen said so. */}
        <p className="mt-1 max-w-[60ch] text-body" style={{ color: 'var(--flow-text-muted)' }}>
          {t('archive.explanation')}
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
        ) : archived.length === 0 ? (
          <EmptyState
            icon={ArchiveIcon}
            title={t('archive.emptyTitle')}
            description={t('archive.emptyDescription')}
            action={{
              label: t('archive.emptyCta'),
              onClick: () => navigate({ to: '/my-issues' }),
            }}
          />
        ) : (
          archived.map((task) => (
            <div
              key={task.id}
              className="flex items-center gap-3 px-6 py-3 hover:bg-[var(--flow-hover)] group"
              style={{ borderBottom: '1px solid var(--flow-border)' }}
            >
              {/* The glyph and title dim together, S-14's 62%. Restore stays
                  outside this wrapper: dimming the row as a whole would take
                  the one control on it down with the content. */}
              <span className="flex min-w-0 flex-1 items-center gap-3 opacity-[0.62]">
                <PriorityIcon priority={task.priority} size={12} />
                <button
                  type="button"
                  onClick={() => openTask(String(task.id))}
                  className="min-w-0 flex-1 truncate text-left text-meta"
                  style={{ color: 'var(--flow-text-muted)', textDecoration: 'line-through' }}
                >
                  {task.title}
                </button>
              </span>
              <span
                className="text-meta shrink-0 hidden group-hover:inline"
                style={{ color: 'var(--flow-text-dim)' }}
              >
                {task.projectName}
              </span>
              <button
                type="button"
                onClick={() =>
                  restoreTask.mutate(
                    { id: task.id, data: { archived: false } },
                    {
                      onSuccess: () => toast.success(t('toasts.task.restored')),
                      onError: () => toast.error(t('toasts.task.restoreFailed')),
                    }
                  )
                }
                disabled={restoreTask.isPending}
                className="flex items-center gap-1 px-2 py-1 rounded-sm text-meta font-medium opacity-0 group-hover:opacity-100 transition-opacity"
                style={{ background: 'var(--flow-hover)', color: 'var(--flow-text-muted)' }}
                title={t('archive.restoreTitle')}
              >
                <RotateCcw size={10} /> {t('archive.restore')}
              </button>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default Archive;
