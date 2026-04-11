import React from 'react';
import { Archive as ArchiveIcon } from 'lucide-react';
import { useFlowStore } from '../../shared/stores/flowStore';
import TaskRow from '../../shared/components/TaskRow';
import EmptyState from '../../shared/components/EmptyState';

const Archive: React.FC = () => {
  const { tasks } = useFlowStore();
  const archived = tasks.filter((t) => t.archived || t.status === 'cancelled');

  return (
    <div className="flex flex-col h-full">
      <div className="px-6 py-4 shrink-0" style={{ borderBottom: '1px solid var(--flow-border)' }}>
        <h1 className="text-[15px] font-semibold" style={{ color: 'var(--flow-text)' }}>
          Archive
        </h1>
        <p className="text-[11px]" style={{ color: 'var(--flow-text-dim)' }}>
          {archived.length} archived {archived.length === 1 ? 'task' : 'tasks'}
        </p>
      </div>
      <div className="flex-1 overflow-y-auto">
        {archived.length === 0 ? (
          <EmptyState
            icon={ArchiveIcon}
            title="Nothing archived"
            description="Archived tasks will show up here."
          />
        ) : (
          archived.map((t) => <TaskRow key={t.id} task={t} showProject />)
        )}
      </div>
    </div>
  );
};

export default Archive;
