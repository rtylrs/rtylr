import React, { useState, useMemo } from 'react';
import { motion } from 'framer-motion';
import { ListTodo, Filter, ArrowUpDown } from 'lucide-react';
import { useMyIssues } from '../../hooks/useMyIssues';
import { useFlowStore, Task, STATUS_LABELS, PRIORITY_ORDER } from '../../stores/flowStore';
import TaskRow from '../../components/TaskRow';
import EmptyState from '../../components/EmptyState';

type GroupBy = 'status' | 'priority' | 'project' | 'none';
type SortBy = 'priority' | 'dueDate' | 'updated';

const MyIssues: React.FC = () => {
  const tasks = useMyIssues();
  const { projects } = useFlowStore();
  const [groupBy, setGroupBy] = useState<GroupBy>('status');
  const [sortBy, setSortBy] = useState<SortBy>('priority');

  const sorted = useMemo(() => {
    const arr = [...tasks];
    arr.sort((a, b) => {
      if (sortBy === 'priority') return PRIORITY_ORDER[a.priority] - PRIORITY_ORDER[b.priority];
      if (sortBy === 'dueDate') return a.dueDate.localeCompare(b.dueDate);
      return a.updatedAt.localeCompare(b.updatedAt);
    });
    return arr;
  }, [tasks, sortBy]);

  const groups = useMemo(() => {
    if (groupBy === 'none') return [{ label: 'All', tasks: sorted }];
    const map: Record<string, Task[]> = {};
    sorted.forEach((t) => {
      let key = 'All';
      if (groupBy === 'status') key = STATUS_LABELS[t.status];
      else if (groupBy === 'priority') key = t.priority;
      else if (groupBy === 'project') key = t.projectName;
      if (!map[key]) map[key] = [];
      map[key].push(t);
    });
    return Object.entries(map).map(([label, tasks]) => ({ label, tasks }));
  }, [sorted, groupBy]);

  return (
    <div className="flex flex-col h-full">
      <div
        className="flex items-center justify-between px-6 py-4 shrink-0"
        style={{ borderBottom: '1px solid var(--flow-border)' }}
      >
        <div>
          <h1 className="text-[15px] font-semibold" style={{ color: 'var(--flow-text)' }}>
            My Issues
          </h1>
          <p className="text-[11px]" style={{ color: 'var(--flow-text-dim)' }}>
            {tasks.length} tasks assigned to you
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
              { value: 'status', label: 'Status' },
              { value: 'priority', label: 'Priority' },
              { value: 'project', label: 'Project' },
              { value: 'none', label: 'No grouping' },
            ]}
          />
        </div>
      </div>

      <div className="flex-1 overflow-y-auto">
        {tasks.length === 0 ? (
          <EmptyState
            icon={ListTodo}
            title="No tasks assigned"
            description="Nothing is on your plate right now. Nice work!"
          />
        ) : (
          groups.map(({ label, tasks: taskList }) => (
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
              </div>
              {taskList.map((t) => (
                <TaskRow key={t.id} task={t} showProject />
              ))}
            </motion.div>
          ))
        )}
      </div>
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
      className="bg-transparent outline-none text-[11px]"
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
