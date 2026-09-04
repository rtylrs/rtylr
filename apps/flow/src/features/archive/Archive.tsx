import React, { useMemo } from 'react';
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

const Archive: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { data: apiTasks = [] } = useTasks({ activeOnly: false });
  const restoreTask = useUpdateTask();

  const archived = useMemo(() => archivedTasks(apiTasks as ApiTask[]), [apiTasks]);

  return (
    <div className="flex flex-col h-full">
      <HelmetContainer title={`${t('archive.title')} · Flow`} />
      <div className="px-6 py-4 shrink-0" style={{ borderBottom: '1px solid var(--flow-border)' }}>
        <h1 className="text-title font-semibold" style={{ color: 'var(--flow-text)' }}>
          {t('archive.title')}
        </h1>
        <p className="text-meta" style={{ color: 'var(--flow-text-dim)' }}>
          {t('archive.count', { count: archived.length })}
        </p>
      </div>
      <div className="flex-1 overflow-y-auto">
        {archived.length === 0 ? (
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
              <PriorityIcon priority={task.priority} size={12} />
              <button
                type="button"
                onClick={() => navigate({ to: `/task/${task.id}` as string })}
                className="flex-1 text-left text-meta truncate"
                style={{ color: 'var(--flow-text-muted)', textDecoration: 'line-through' }}
              >
                {task.title}
              </button>
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
