import React, { useState } from 'react';
import { Circle, ChevronDown } from 'lucide-react';
import { useFlowStore } from '../stores/flowStore';

interface ProjectSelectProps {
  value?: string;
  onChange: (projectId: string) => void;
  placeholder?: string;
}

export const ProjectSelect: React.FC<ProjectSelectProps> = ({
  value,
  onChange,
  placeholder = 'Select project',
}) => {
  const [open, setOpen] = useState(false);
  const { projects } = useFlowStore();
  const current = projects.find((p) => p.id === value);

  return (
    <div className="relative">
      <button
        type="button"
        onClick={() => setOpen((o) => !o)}
        className="flex items-center gap-2 px-2.5 py-1.5 rounded-sm text-meta transition-colors hover:bg-[var(--flow-hover)] w-full"
        style={{
          color: current ? 'var(--flow-text)' : 'var(--flow-text-muted)',
          background: 'var(--flow-input-bg)',
          border: '1px solid var(--flow-border-control)',
        }}
      >
        {current ? (
          <>
            <Circle size={8} className="fill-current" style={{ color: current.color }} />
            <span className="flex-1 text-start truncate">{current.name}</span>
          </>
        ) : (
          <span className="flex-1 text-start">{placeholder}</span>
        )}
        <ChevronDown size={11} />
      </button>
      {open && (
        <div
          className="absolute top-full start-0 mt-1 rounded-lg py-1 z-50 min-w-full max-h-[240px] overflow-y-auto"
          style={{
            background: 'var(--flow-card)',
            border: '1px solid var(--flow-border)',
            boxShadow: 'var(--flow-shadow)',
          }}
        >
          {projects.map((p) => (
            <button
              type="button"
              key={p.id}
              onClick={() => {
                onChange(p.id);
                setOpen(false);
              }}
              className="w-full flex items-center gap-2 px-3 py-1.5 text-meta hover:bg-[var(--flow-hover)] text-start"
              style={{ color: 'var(--flow-text)' }}
            >
              <Circle size={8} className="fill-current" style={{ color: p.color }} />
              {p.name}
            </button>
          ))}
        </div>
      )}
    </div>
  );
};

export default ProjectSelect;
