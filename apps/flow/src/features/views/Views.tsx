import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { Eye, Plus, Flame, Timer, CheckCircle2 } from 'lucide-react';
import { useFlowStore, SavedView } from '../../shared/stores/flowStore';
import { useViews, useFilteredTasks } from './useViews';
import TaskRow from '../../shared/components/TaskRow';

const ICON_MAP: Record<string, React.ComponentType<{ size?: number }>> = {
  flame: Flame,
  timer: Timer,
  'check-circle': CheckCircle2,
};

const Views: React.FC = () => {
  const { data: views = [] } = useViews();
  const { tasks: allTasks } = useFlowStore();
  const [selected, setSelected] = useState<SavedView | null>(null);
  const displayView = selected || views[0] || null;
  const tasks = useFilteredTasks(displayView, allTasks);

  return (
    <div className="flex h-full">
      <aside
        className="w-64 shrink-0 px-3 py-4"
        style={{
          background: 'var(--flow-surface)',
          borderRight: '1px solid var(--flow-border)',
        }}
      >
        <div className="flex items-center justify-between px-2 mb-3">
          <h2 className="text-[11px] font-semibold uppercase tracking-wider"
            style={{ color: 'var(--flow-text-muted)' }}>
            Saved Views
          </h2>
          <button
            className="h-5 w-5 rounded flex items-center justify-center hover:bg-[var(--flow-hover)]"
            style={{ color: 'var(--flow-text-dim)' }}
          >
            <Plus size={11} />
          </button>
        </div>
        <div className="space-y-0.5">
          {views.map((v) => {
            const Icon = ICON_MAP[v.icon] || Eye;
            const active = selected?.id === v.id;
            return (
              <button
                key={v.id}
                onClick={() => setSelected(v)}
                className="w-full flex items-center gap-2 px-3 py-1.5 rounded text-[12px] text-left"
                style={{
                  background: active ? 'var(--flow-active)' : 'transparent',
                  color: active ? 'var(--flow-text)' : 'var(--flow-text-muted)',
                }}
              >
                <Icon size={12} />
                {v.name}
              </button>
            );
          })}
        </div>
      </aside>
      <div className="flex-1 overflow-y-auto">
        <div className="px-6 py-4"
          style={{ borderBottom: '1px solid var(--flow-border)' }}>
          <h1 className="text-[15px] font-semibold" style={{ color: 'var(--flow-text)' }}>
            {selected?.name || 'Pick a view'}
          </h1>
          <p className="text-[11px]" style={{ color: 'var(--flow-text-dim)' }}>
            {tasks.length} task{tasks.length !== 1 ? 's' : ''}
          </p>
        </div>
        <motion.div
          key={selected?.id}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.18 }}
        >
          {tasks.map((t) => (
            <TaskRow key={t.id} task={t} showProject />
          ))}
        </motion.div>
      </div>
    </div>
  );
};

export default Views;
