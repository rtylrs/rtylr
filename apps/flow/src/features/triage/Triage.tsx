import React, { useMemo } from 'react';
import { Eye } from 'lucide-react';
import type { Task } from '../../shared/stores/flowStore';
import { useFlowStore } from '../../shared/stores/flowStore';
import { HelmetContainer } from '@voxire/components';
import { useTasks } from '../tasks/useTasks';
import TaskRow from '../../shared/components/TaskRow';
import EmptyState from '../../shared/components/EmptyState';
import { apiTaskToTask } from '../../shared/lib/taskMapper';

const Triage: React.FC = () => {
  const { data: apiTasks } = useTasks({});
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
      <HelmetContainer title="Triage · rtylr" />
      <div className="px-6 py-4 shrink-0" style={{ borderBottom: '1px solid var(--flow-border)' }}>
        <h1 className="text-[15px] font-semibold" style={{ color: 'var(--flow-text)' }}>
          Triage
        </h1>
        <p className="text-[11px]" style={{ color: 'var(--flow-text-dim)' }}>
          {needsTriage.length} tasks need attention · Capture → set priority & date → move out of
          backlog
        </p>
      </div>
      <div className="flex-1 overflow-y-auto">
        {needsTriage.length === 0 ? (
          <EmptyState
            icon={Eye}
            title="Inbox zero"
            description="Nothing stuck without a priority or due date. New work lands in capture first-keep triage lean so planned work stays trustworthy."
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
