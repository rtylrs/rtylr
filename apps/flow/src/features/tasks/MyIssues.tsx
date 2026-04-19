import React, { useMemo, useState } from 'react';
import { motion } from 'motion/react';
import {
  ListTodo,
  Filter,
  ArrowUpDown,
  Sun,
  CalendarDays,
  CalendarRange,
  Layers,
  Plus,
} from 'lucide-react';
import { useMyIssues } from './useMyIssues';
import { useCreateTask } from './useTasks';
import { Task, STATUS_LABELS, PRIORITY_ORDER, useFlowStore } from '../../shared/stores/flowStore';
import { apiTaskToTask } from '../../shared/lib/taskMapper';
import TaskRow from '../../shared/components/TaskRow';
import EmptyState from '../../shared/components/EmptyState';
import {
  bucketForMyTask,
  MY_TASK_BUCKET_LABEL,
  MY_TASK_BUCKET_ORDER,
  type MyTaskBucket,
} from '../../shared/lib/myTasksBuckets';

/** `schedule` = grouped by due window (Recently assigned, Today, Upcoming, Later). */
type GroupBy = 'schedule' | 'status' | 'priority' | 'project' | 'none';
type SortBy = 'priority' | 'dueDate' | 'updated';

const BUCKET_ICONS: Record<MyTaskBucket, React.ReactNode> = {
  recently_assigned: <Layers size={12} />,
  today: <Sun size={12} />,
  upcoming: <CalendarDays size={12} />,
  later: <CalendarRange size={12} />,
};

const MyIssues: React.FC = () => {
  const { data: rawTasks = [], isLoading } = useMyIssues();
  const tasks = rawTasks.map(apiTaskToTask);
  const { selectedTaskIds } = useFlowStore();
  const [groupBy, setGroupBy] = useState<GroupBy>('schedule');
  const [sortBy, setSortBy] = useState<SortBy>('priority');

  const sorted = useMemo(() => {
    const arr = [...tasks];
    arr.sort((a, b) => {
      if (sortBy === 'priority') return PRIORITY_ORDER[a.priority] - PRIORITY_ORDER[b.priority];
      if (sortBy === 'dueDate') return (a.dueDate || '').localeCompare(b.dueDate || '');
      return a.updatedAt.localeCompare(b.updatedAt);
    });
    return arr;
  }, [tasks, sortBy]);

  const groups = useMemo(() => {
    if (groupBy === 'schedule') {
      return MY_TASK_BUCKET_ORDER.map((bucket) => ({
        label: MY_TASK_BUCKET_LABEL[bucket],
        icon: BUCKET_ICONS[bucket],
        tasks: sorted.filter((t) => bucketForMyTask(t) === bucket),
        bucket,
      }));
    }
    if (groupBy === 'none')
      return [{ label: 'All', icon: null, tasks: sorted, bucket: null as MyTaskBucket | null }];
    const map: Record<string, Task[]> = {};
    sorted.forEach((t) => {
      let key = 'All';
      if (groupBy === 'status') key = STATUS_LABELS[t.status];
      else if (groupBy === 'priority') key = t.priority;
      else if (groupBy === 'project') key = t.projectName;
      if (!map[key]) map[key] = [];
      map[key].push(t);
    });
    return Object.entries(map).map(([label, taskList]) => ({
      label,
      icon: null,
      tasks: taskList,
      bucket: null as MyTaskBucket | null,
    }));
  }, [sorted, groupBy]);

  return (
    <div className="flex flex-col h-full">
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
        <div className="flex items-center gap-2">
          <ToolbarSelect
            icon={<ArrowUpDown size={11} />}
            label="Sort"
            value={sortBy}
            onChange={(v) => setSortBy(v as SortBy)}
            options={[
              { value: 'priority', label: 'Priority' },
              { value: 'dueDate', label: 'Due date' },
              { value: 'updated', label: 'Updated' },
            ]}
          />
          <ToolbarSelect
            icon={<Filter size={11} />}
            label="Group"
            value={groupBy}
            onChange={(v) => setGroupBy(v as GroupBy)}
            options={[
              { value: 'schedule', label: 'Schedule' },
              { value: 'status', label: 'Status' },
              { value: 'priority', label: 'Priority' },
              { value: 'project', label: 'Project' },
              { value: 'none', label: 'None' },
            ]}
          />
        </div>
      </div>

      <QuickAddRow />

      <div className="flex-1 overflow-y-auto">
        {tasks.length === 0 ? (
          <EmptyState
            icon={ListTodo}
            title="No tasks assigned"
            description="Nothing is on your plate right now. Nice work!"
          />
        ) : (
          groups.map(({ label, tasks: taskList, icon, bucket }) => (
            <motion.div
              key={label}
              initial={{ opacity: 0, y: 4 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.15 }}
            >
              <div
                className="px-6 py-2 flex items-center gap-2 sticky top-0 z-10"
                style={{
                  background: 'var(--flow-bg)',
                  borderBottom: '1px solid var(--flow-border)',
                }}
              >
                {icon && (
                  <span style={{ color: 'var(--flow-text-muted)' }} className="flex items-center">
                    {icon}
                  </span>
                )}
                <span
                  className="text-[11px] font-semibold uppercase tracking-wider"
                  style={{ color: 'var(--flow-text-muted)' }}
                >
                  {label}
                </span>
                <span
                  className="text-[10px] px-1.5 py-0.5 rounded"
                  style={{
                    background: 'var(--flow-hover)',
                    color: 'var(--flow-text-dim)',
                  }}
                >
                  {taskList.length}
                </span>
                {groupBy === 'schedule' && bucket === 'today' && (
                  <span className="text-[10px] ml-1" style={{ color: 'var(--flow-text-dim)' }}>
                    Includes overdue
                  </span>
                )}
              </div>
              {taskList.map((t) => (
                <TaskRow
                  key={t.id}
                  task={t}
                  showProject
                  selected={selectedTaskIds.includes(t.id)}
                />
              ))}
            </motion.div>
          ))
        )}
      </div>
    </div>
  );
};

/**
 * Inline row that lets the user type a task title, hit Enter, and see it
 * appear in their list. The backend auto-resolves the caller's Inbox
 * project inside their Personal workspace — no project picker needed.
 *
 * react-query invalidation on `['my-tasks']` pulls the new task into the
 * list after the POST resolves.
 */
const QuickAddRow: React.FC = () => {
  const [title, setTitle] = useState('');
  const createTask = useCreateTask();

  const submit = (): void => {
    const trimmed = title.trim();
    if (!trimmed) return;
    // workspaceId + projectId intentionally omitted — the backend
    // resolves "personal" workspace + the Inbox project automatically.
    createTask.mutate(
      {
        workspaceId: '',
        projectId: '',
        title: trimmed,
      },
      {
        onSuccess: () => setTitle(''),
      }
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
        placeholder="Add a task and press Enter — lands in your Inbox"
        disabled={createTask.isPending}
        className="flex-1 bg-transparent outline-none text-[13px] placeholder:opacity-50"
        style={{ color: 'var(--flow-text)' }}
      />
      {title.trim().length > 0 && (
        <button
          onClick={submit}
          disabled={createTask.isPending}
          className="text-[11px] font-semibold px-2.5 py-1 rounded transition-opacity hover:opacity-90 disabled:opacity-50"
          style={{ background: '#64C5B9', color: '#ffffff' }}
        >
          Add
        </button>
      )}
    </div>
  );
};

const ToolbarSelect: React.FC<{
  icon: React.ReactNode;
  label: string;
  value: string;
  onChange: (v: string) => void;
  options: { value: string; label: string }[];
}> = ({ icon, label, value, onChange, options }) => (
  <label
    className="flex items-center gap-1.5 px-2.5 py-1 rounded text-[11px]"
    style={{
      background: 'var(--flow-input-bg)',
      border: '1px solid var(--flow-border)',
      color: 'var(--flow-text-muted)',
    }}
  >
    {icon}
    <span>{label}:</span>
    <select
      value={value}
      onChange={(e) => onChange(e.target.value)}
      className="bg-transparent outline-none text-[11px] max-w-[160px]"
      style={{ color: 'var(--flow-text)' }}
    >
      {options.map((o) => (
        <option key={o.value} value={o.value} style={{ background: 'var(--flow-card)' }}>
          {o.label}
        </option>
      ))}
    </select>
  </label>
);

export default MyIssues;
