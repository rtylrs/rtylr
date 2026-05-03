import React, { useState, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { Inbox, Plus } from 'lucide-react';
import { useFlowStore } from '../../shared/stores/flowStore';
import { useTasks, useUpdateTaskStatus } from '../tasks/useTasks';
import PriorityIcon from '../../shared/components/PriorityIcon';
import StatusIcon from '../../shared/components/StatusIcon';
import type { ApiTask } from '../../shared/services/flowApi';

const Backlog: React.FC = () => {
  const navigate = useNavigate();
  const { projects, setQuickCreateOpen } = useFlowStore();
  const { data: apiTasks = [] } = useTasks({});
  const updateStatus = useUpdateTaskStatus();
  const [filter, setFilter] = useState<'all' | 'todo' | 'backlog'>('all');

  const backlogTasks = useMemo(() => {
    return apiTasks
      .filter((t) => {
        const s = t.status;
        if (filter === 'all') return s === 'todo' || s === 'backlog';
        return s === filter;
      })
      .sort((a, b) => {
        const pOrder: Record<string, number> = { urgent: 0, high: 1, medium: 2, low: 3 };
        return (pOrder[a.priority] ?? 4) - (pOrder[b.priority] ?? 4);
      });
  }, [apiTasks, filter]);

  const byProject = useMemo(() => {
    const map: Record<string, ApiTask[]> = {};
    for (const t of backlogTasks) {
      if (!map[t.projectId]) map[t.projectId] = [];
      map[t.projectId].push(t);
    }
    return map;
  }, [backlogTasks]);

  return (
    <div className="flex flex-col h-full">
      <div
        className="flex items-center justify-between px-6 py-4 shrink-0"
        style={{ borderBottom: '1px solid var(--flow-border)' }}
      >
        <div className="flex items-center gap-2">
          <Inbox size={14} style={{ color: 'var(--flow-text-muted)' }} />
          <h1 className="text-[15px] font-semibold" style={{ color: 'var(--flow-text)' }}>
            Backlog
          </h1>
          <span
            className="text-[11px] px-2 py-0.5 rounded-full font-semibold"
            style={{ background: 'var(--flow-hover)', color: 'var(--flow-text-dim)' }}
          >
            {backlogTasks.length}
          </span>
        </div>
        <div className="flex items-center gap-2">
          <div
            className="flex items-center rounded-lg p-0.5"
            style={{ background: 'var(--flow-hover)', border: '1px solid var(--flow-border)' }}
          >
            {(['all', 'todo', 'backlog'] as const).map((f) => (
              <button
                key={f}
                onClick={() => setFilter(f)}
                className="px-2.5 py-1 rounded-md text-[11px] font-medium transition-colors"
                style={{
                  background: filter === f ? 'var(--flow-card)' : 'transparent',
                  color: filter === f ? 'var(--flow-text)' : 'var(--flow-text-dim)',
                }}
              >
                {f.charAt(0).toUpperCase() + f.slice(1).replace('_', ' ')}
              </button>
            ))}
          </div>
          <button
            onClick={() => setQuickCreateOpen(true)}
            className="flex items-center gap-1.5 h-7 px-2.5 rounded text-[11px] font-semibold"
            style={{ background: 'var(--accent)', color: '#0a0a0a' }}
          >
            <Plus size={11} /> New task
          </button>
        </div>
      </div>

      <div className="flex-1 overflow-auto p-6 space-y-6">
        {backlogTasks.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-48 gap-2">
            <Inbox size={28} style={{ color: 'var(--flow-text-dim)', opacity: 0.4 }} />
            <p className="text-[13px]" style={{ color: 'var(--flow-text-dim)' }}>
              Backlog is empty
            </p>
          </div>
        ) : (
          Object.entries(byProject).map(([projectId, projectTasks]) => {
            const project = projects.find((p) => p.id === projectId);
            return (
              <div key={projectId}>
                <div className="flex items-center gap-2 mb-2">
                  <span
                    className="h-2 w-2 rounded-full"
                    style={{ background: project?.color ?? '#6b7280' }}
                  />
                  <span className="text-[12px] font-semibold" style={{ color: 'var(--flow-text)' }}>
                    {project?.name ?? projectTasks[0]?.projectName ?? 'Unknown project'}
                  </span>
                  <span className="text-[10px]" style={{ color: 'var(--flow-text-dim)' }}>
                    {projectTasks.length}
                  </span>
                </div>
                <div
                  className="rounded-lg overflow-hidden"
                  style={{ border: '1px solid var(--flow-border)' }}
                >
                  {projectTasks.map((task, i) => (
                    <button
                      key={task.id}
                      onClick={() => navigate(`/task/${task.id}`)}
                      className="w-full flex items-center gap-3 px-4 py-3 text-left hover:bg-[var(--flow-hover)] transition-colors"
                      style={{
                        borderTop: i > 0 ? '1px solid var(--flow-border)' : 'none',
                      }}
                    >
                      <PriorityIcon priority={task.priority} size={12} />
                      <StatusIcon status={task.status} size={12} />
                      <span
                        className="flex-1 text-[12px] truncate"
                        style={{ color: 'var(--flow-text)' }}
                      >
                        {task.title}
                      </span>
                      {task.assignee && (
                        <span
                          className="h-5 w-5 rounded-full flex items-center justify-center text-[9px] font-bold shrink-0"
                          style={{ background: 'rgba(100,197,185,0.2)', color: '#64C5B9' }}
                          title={task.assignee}
                        >
                          {task.assigneeInitials || task.assignee[0]}
                        </span>
                      )}
                      {task.dueDate && (
                        <span
                          className="text-[10px] shrink-0"
                          style={{ color: 'var(--flow-text-dim)' }}
                        >
                          {task.dueDate}
                        </span>
                      )}
                      <select
                        value={task.status}
                        onClick={(e) => e.stopPropagation()}
                        onChange={(e) =>
                          updateStatus.mutate({
                            id: task.id,
                            status: e.target.value as ApiTask['status'],
                          })
                        }
                        className="text-[10px] rounded px-1 py-0.5 outline-none shrink-0"
                        style={{
                          background: 'var(--flow-hover)',
                          border: '1px solid var(--flow-border)',
                          color: 'var(--flow-text-muted)',
                        }}
                      >
                        <option value="backlog">Backlog</option>
                        <option value="todo">Todo</option>
                        <option value="in_progress">In Progress</option>
                        <option value="in_review">In Review</option>
                        <option value="done">Done</option>
                      </select>
                    </button>
                  ))}
                </div>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
};

export default Backlog;
