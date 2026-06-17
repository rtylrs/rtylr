import React, { useMemo, useState } from 'react';
import { toast } from 'sonner';
import { motion } from 'motion/react';
import { ListTodo, Plus } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { useMyIssues } from './useMyIssues';
import { useCreateTask } from './useTasks';
import { useFlowStore } from '../../shared/stores/flowStore';
import { apiTaskToTask } from '../../shared/lib/taskMapper';
import TaskCard from '../../shared/components/TaskCard';
import EmptyState from '../../shared/components/EmptyState';

// Backlog lives here as the leftmost column instead of as its own
// sidebar tab - keeping unscheduled work next to active work means a
// single view covers personal task management end-to-end.
const KANBAN_COLUMNS: { status: string; label: string; color: string }[] = [
  { status: 'backlog', label: 'Backlog', color: '#6b7280' },
  { status: 'todo', label: 'Todo', color: '#9ca3af' },
  { status: 'in_progress', label: 'In Progress', color: '#3b82f6' },
  { status: 'in_review', label: 'In Review', color: '#f59e0b' },
  { status: 'done', label: 'Done', color: '#22c55e' },
];

const MyIssues: React.FC = () => {
  const { data: rawTasks = [], isLoading } = useMyIssues();
  const tasks = useMemo(() => rawTasks.map(apiTaskToTask), [rawTasks]);
  const { selectedTaskIds } = useFlowStore();

  const byStatus = useMemo(() => {
    const map: Record<string, typeof tasks> = {};
    for (const col of KANBAN_COLUMNS) map[col.status] = [];
    for (const t of tasks) {
      if (map[t.status] !== undefined) map[t.status].push(t);
    }
    return map;
  }, [tasks]);

  return (
    <div className="flex flex-col h-full">
      <HelmetContainer title="My tasks · rtylr" />
      <div
        className="flex items-center justify-between px-6 py-4 shrink-0"
        style={{ borderBottom: '1px solid var(--flow-border)' }}
      >
        <div>
          <h1 className="text-[15px] font-semibold" style={{ color: 'var(--flow-text)' }}>
            My tasks
          </h1>
          <p className="text-[11px]" style={{ color: 'var(--flow-text-dim)' }}>
            {isLoading ? 'Loading…' : `${tasks.length} tasks assigned to you`}
          </p>
        </div>
      </div>

      <QuickAddRow />

      {tasks.length === 0 && !isLoading ? (
        <EmptyState
          icon={ListTodo}
          title="No tasks assigned"
          description="Nothing is on your plate right now. Nice work!"
        />
      ) : (
        <div className="flex-1 overflow-x-auto overflow-y-hidden">
          <div className="flex gap-3 h-full p-4 min-w-max">
            {KANBAN_COLUMNS.map((col, colIdx) => {
              const colTasks = byStatus[col.status] ?? [];
              return (
                <motion.div
                  key={col.status}
                  initial={{ opacity: 0, y: 8 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.15, delay: colIdx * 0.04 }}
                  className="flex flex-col rounded-xl shrink-0"
                  style={{
                    width: 280,
                    background: 'var(--flow-card)',
                    border: '1px solid var(--flow-border)',
                  }}
                >
                  {/* Column header */}
                  <div
                    className="flex items-center gap-2 px-3 py-2.5 shrink-0"
                    style={{ borderBottom: '1px solid var(--flow-border)' }}
                  >
                    <span
                      className="h-2 w-2 rounded-full shrink-0"
                      style={{ background: col.color }}
                    />
                    <span
                      className="text-[12px] font-semibold flex-1"
                      style={{ color: 'var(--flow-text)' }}
                    >
                      {col.label}
                    </span>
                    <span
                      className="text-[10px] px-1.5 py-0.5 rounded font-medium"
                      style={{
                        background: `${col.color}1a`,
                        color: col.color,
                      }}
                    >
                      {colTasks.length}
                    </span>
                  </div>

                  {/* Tasks */}
                  <div className="flex-1 overflow-y-auto p-2 space-y-1.5">
                    {colTasks.length === 0 ? (
                      <p
                        className="text-center text-[11px] py-6"
                        style={{ color: 'var(--flow-text-dim)' }}
                      >
                        No tasks
                      </p>
                    ) : (
                      colTasks.map((t) => (
                        <TaskCard
                          key={t.id}
                          task={t}
                          compact
                          showProject
                          selected={selectedTaskIds.includes(t.id)}
                        />
                      ))
                    )}
                  </div>
                </motion.div>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
};

const QuickAddRow: React.FC = () => {
  const [title, setTitle] = useState('');
  const createTask = useCreateTask();

  const submit = (): void => {
    const trimmed = title.trim();
    if (!trimmed) {
      toast.error('Task title is required');
      return;
    }
    // Personal tasks land in Backlog so the leftmost column actually
    // collects new captures instead of jamming them into Todo.
    createTask.mutate(
      { workspaceId: '', projectId: '', title: trimmed, status: 'backlog' },
      { onSuccess: () => setTitle('') }
    );
  };

  return (
    <div
      className="flex items-center gap-2 px-6 py-2 shrink-0"
      style={{ borderBottom: '1px solid var(--flow-border)' }}
    >
      <Plus size={12} style={{ color: 'var(--flow-text-dim)' }} />
      <input
        value={title}
        onChange={(e) => setTitle(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === 'Enter') {
            e.preventDefault();
            submit();
          }
        }}
        placeholder="Add a personal task and press Enter"
        disabled={createTask.isPending}
        className="flex-1 bg-transparent outline-none text-[13px] placeholder:opacity-50"
        style={{ color: 'var(--flow-text)' }}
      />
      {title.trim().length > 0 && (
        <button
          onClick={submit}
          disabled={createTask.isPending}
          className="text-[11px] font-semibold px-2.5 py-1 rounded transition-opacity hover:opacity-90 disabled:opacity-50"
          style={{ background: 'var(--accent)', color: '#ffffff' }}
        >
          Add
        </button>
      )}
    </div>
  );
};

export default MyIssues;
