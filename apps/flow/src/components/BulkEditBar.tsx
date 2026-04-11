import React from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { toast } from 'sonner';
import { X, Check, Trash2, Archive, Flag } from 'lucide-react';
import { useFlowStore, Task } from '../stores/flowStore';

export const BulkEditBar: React.FC = () => {
  const { selectedTaskIds, clearSelection, bulkUpdateTasks, deleteTask } = useFlowStore();

  if (selectedTaskIds.length === 0) return null;

  const act = (patch: Partial<Task>) => {
    bulkUpdateTasks(selectedTaskIds, patch);
    toast.success(`Updated ${selectedTaskIds.length} task${selectedTaskIds.length !== 1 ? 's' : ''}`);
    clearSelection();
  };

  return (
    <AnimatePresence>
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        exit={{ opacity: 0, y: 20 }}
        className="fixed left-1/2 z-50 flex max-w-[calc(100vw-1.5rem)] -translate-x-1/2 items-center gap-1 overflow-x-auto rounded-xl px-3 py-2"
        style={{
          bottom: 'max(1.5rem, env(safe-area-inset-bottom))',
          background: 'var(--flow-card)',
          border: '1px solid var(--flow-border)',
          boxShadow: 'var(--flow-shadow)',
        }}
      >
        <span
          className="text-[11px] font-semibold px-2"
          style={{ color: 'var(--flow-text)' }}
        >
          {selectedTaskIds.length} selected
        </span>
        <div
          className="h-4 w-px mx-1"
          style={{ background: 'var(--flow-border)' }}
        />
        <button
          onClick={() => act({ status: 'done' })}
          className="h-7 px-2 rounded text-[11px] flex items-center gap-1 hover:bg-[var(--flow-hover)]"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          <Check size={11} /> Mark done
        </button>
        <button
          onClick={() => act({ priority: 'urgent' })}
          className="h-7 px-2 rounded text-[11px] flex items-center gap-1 hover:bg-[var(--flow-hover)]"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          <Flag size={11} /> Set urgent
        </button>
        <button
          onClick={() => {
            selectedTaskIds.forEach((id) => useFlowStore.getState().archiveTask(id));
            toast.success(`Archived ${selectedTaskIds.length} tasks`);
            clearSelection();
          }}
          className="h-7 px-2 rounded text-[11px] flex items-center gap-1 hover:bg-[var(--flow-hover)]"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          <Archive size={11} /> Archive
        </button>
        <button
          onClick={() => {
            selectedTaskIds.forEach((id) => deleteTask(id));
            toast.success(`Deleted ${selectedTaskIds.length} tasks`);
            clearSelection();
          }}
          className="h-7 px-2 rounded text-[11px] flex items-center gap-1 hover:bg-[var(--flow-hover)]"
          style={{ color: '#ef4444' }}
        >
          <Trash2 size={11} /> Delete
        </button>
        <div
          className="h-4 w-px mx-1"
          style={{ background: 'var(--flow-border)' }}
        />
        <button
          onClick={clearSelection}
          className="h-7 w-7 rounded flex items-center justify-center hover:bg-[var(--flow-hover)]"
          style={{ color: 'var(--flow-text-dim)' }}
        >
          <X size={12} />
        </button>
      </motion.div>
    </AnimatePresence>
  );
};

export default BulkEditBar;
