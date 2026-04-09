import React, { useMemo } from 'react';
import { Eye } from 'lucide-react';
import { useFlowStore } from '../../stores/flowStore';
import TaskRow from '../../components/TaskRow';
import EmptyState from '../../components/EmptyState';

const Triage: React.FC = () => {
  const { tasks } = useFlowStore();

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
      <div
        className="px-6 py-4 shrink-0"
        style={{ borderBottom: '1px solid var(--flow-border)' }}
      >
        <h1 className="text-[15px] font-semibold" style={{ color: 'var(--flow-text)' }}>
          Triage
        </h1>
        <p className="text-[11px]" style={{ color: 'var(--flow-text-dim)' }}>
          {needsTriage.length} tasks need attention
        </p>
      </div>
      <div className="flex-1 overflow-y-auto">
        {needsTriage.length === 0 ? (
          <EmptyState
            icon={Eye}
            title="Triage inbox is empty"
            description="All incoming tasks have been prioritized or routed."
          />
        ) : (
          needsTriage.map((t) => <TaskRow key={t.id} task={t} showProject />)
        )}
      </div>
    </div>
  );
};

export default Triage;
