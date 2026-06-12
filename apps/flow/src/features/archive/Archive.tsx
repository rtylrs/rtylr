import React, { useMemo } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { Archive as ArchiveIcon, RotateCcw } from 'lucide-react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { useTasks, useUpdateTaskStatus } from '../tasks/useTasks';
import type { ApiTask } from '../../shared/services/flowApi';
import PriorityIcon from '../../shared/components/PriorityIcon';
import EmptyState from '../../shared/components/EmptyState';

const Archive: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { data: apiTasks = [] } = useTasks({});
  const restoreTask = useUpdateTaskStatus();

  const archived = useMemo(
    () => (apiTasks as ApiTask[]).filter((task) => task.status === 'cancelled'),
    [apiTasks]
  );

  return (
    <div className="flex flex-col h-full">
      <HelmetContainer title="Archive · rtylr" />
      <div className="px-6 py-4 shrink-0" style={{ borderBottom: '1px solid var(--flow-border)' }}>
        <h1 className="text-[15px] font-semibold" style={{ color: 'var(--flow-text)' }}>
          Archive
        </h1>
        <p className="text-[11px]" style={{ color: 'var(--flow-text-dim)' }}>
          {archived.length} archived {archived.length === 1 ? 'task' : 'tasks'} · Cancelled tasks
          live here
        </p>
      </div>
      <div className="flex-1 overflow-y-auto">
        {archived.length === 0 ? (
          <EmptyState
            icon={ArchiveIcon}
            title="Nothing archived"
            description="Cancelled tasks will show up here."
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
                onClick={() => navigate({ to: `/task/${task.id}` as string })}
                className="flex-1 text-left text-[12px] truncate"
                style={{ color: 'var(--flow-text-muted)', textDecoration: 'line-through' }}
              >
                {task.title}
              </button>
              <span
                className="text-[10px] shrink-0 hidden group-hover:inline"
                style={{ color: 'var(--flow-text-dim)' }}
              >
                {task.projectName}
              </span>
              <button
                onClick={() =>
                  restoreTask.mutate(
                    { id: task.id, status: 'todo' },
                    {
                      onSuccess: () => toast.success(t('toasts.task.restored')),
                      onError: () => toast.error(t('toasts.task.restoreFailed')),
                    }
                  )
                }
                disabled={restoreTask.isPending}
                className="flex items-center gap-1 px-2 py-1 rounded text-[10px] font-medium opacity-0 group-hover:opacity-100 transition-opacity"
                style={{ background: 'var(--flow-hover)', color: 'var(--flow-text-muted)' }}
                title="Restore to Todo"
              >
                <RotateCcw size={10} /> Restore
              </button>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default Archive;
