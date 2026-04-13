import React, { useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { Archive as ArchiveIcon, RotateCcw } from 'lucide-react';
import { toast } from 'sonner';
import { useTasks, useUpdateTaskStatus } from '../tasks/useTasks';
import type { ApiTask } from '../../shared/services/flowApi';
import PriorityIcon from '../../shared/components/PriorityIcon';
import EmptyState from '../../shared/components/EmptyState';

const Archive: React.FC = () => {
  const navigate = useNavigate();
  const { data: apiTasks = [] } = useTasks({});
  const restoreTask = useUpdateTaskStatus();

  const archived = useMemo(
    () => (apiTasks as ApiTask[]).filter((t) => t.status === 'cancelled'),
    [apiTasks]
  );

  return (
    <div className="flex flex-col h-full">
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
          archived.map((t) => (
            <div
              key={t.id}
              className="flex items-center gap-3 px-6 py-3 hover:bg-[var(--flow-hover)] group"
              style={{ borderBottom: '1px solid var(--flow-border)' }}
            >
              <PriorityIcon priority={t.priority} size={12} />
              <button
                onClick={() => navigate(`/task/${t.id}`)}
                className="flex-1 text-left text-[12px] truncate"
                style={{ color: 'var(--flow-text-muted)', textDecoration: 'line-through' }}
              >
                {t.title}
              </button>
              <span
                className="text-[10px] shrink-0 hidden group-hover:inline"
                style={{ color: 'var(--flow-text-dim)' }}
              >
                {t.projectName}
              </span>
              <button
                onClick={() =>
                  restoreTask.mutate(
                    { id: t.id, status: 'todo' },
                    {
                      onSuccess: () => toast.success('Task restored'),
                      onError: () => toast.error('Failed to restore task'),
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
