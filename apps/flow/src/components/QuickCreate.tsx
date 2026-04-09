import React, { useEffect, useRef, useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { toast } from 'sonner';
import { Plus, X } from 'lucide-react';
import { useFlowStore, Task } from '../stores/flowStore';
import ProjectSelect from './ProjectSelect';
import PriorityIcon from './PriorityIcon';

export const QuickCreate: React.FC = () => {
  const { quickCreateOpen, setQuickCreateOpen, projects, activeWorkspaceId, createTask } = useFlowStore();
  const [title, setTitle] = useState('');
  const [projectId, setProjectId] = useState<string>('');
  const [priority, setPriority] = useState<Task['priority']>('medium');
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (quickCreateOpen) {
      setTitle('');
      const first = projects.find((p) => p.workspaceId === activeWorkspaceId) || projects[0];
      setProjectId(first?.id || '');
      setPriority('medium');
      setTimeout(() => inputRef.current?.focus(), 10);
    }
  }, [quickCreateOpen, projects, activeWorkspaceId]);

  const submit = () => {
    if (!title.trim() || !projectId) return;
    const project = projects.find((p) => p.id === projectId);
    if (!project) return;
    const task = createTask({
      title: title.trim(),
      projectId,
      workspaceId: project.workspaceId,
      priority,
    });
    toast.success(`Created ${task.identifier}`, { description: task.title });
    setQuickCreateOpen(false);
  };

  const priorities: Task['priority'][] = ['urgent', 'high', 'medium', 'low', 'no_priority'];

  return (
    <AnimatePresence>
      {quickCreateOpen && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.12 }}
          onClick={() => setQuickCreateOpen(false)}
          className="fixed inset-0 z-[100] flex items-start justify-center pt-[15vh] px-4"
          style={{ background: 'rgba(0,0,0,0.4)', backdropFilter: 'blur(4px)' }}
        >
          <motion.div
            initial={{ opacity: 0, y: -10, scale: 0.98 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: -10, scale: 0.98 }}
            transition={{ duration: 0.14 }}
            onClick={(e) => e.stopPropagation()}
            className="w-full max-w-lg rounded-xl overflow-hidden"
            style={{
              background: 'var(--flow-card)',
              border: '1px solid var(--flow-border)',
              boxShadow: 'var(--flow-shadow)',
            }}
          >
            <div
              className="flex items-center justify-between px-4 py-3"
              style={{ borderBottom: '1px solid var(--flow-border)' }}
            >
              <div className="flex items-center gap-2">
                <Plus size={14} style={{ color: '#64C5B9' }} />
                <span
                  className="text-[13px] font-semibold"
                  style={{ color: 'var(--flow-text)' }}
                >
                  New task
                </span>
              </div>
              <button
                onClick={() => setQuickCreateOpen(false)}
                className="h-6 w-6 rounded flex items-center justify-center hover:bg-[var(--flow-hover)]"
                style={{ color: 'var(--flow-text-dim)' }}
              >
                <X size={12} />
              </button>
            </div>
            <div className="p-4 space-y-3">
              <input
                ref={inputRef}
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' && (e.metaKey || e.ctrlKey || !e.shiftKey)) submit();
                  if (e.key === 'Escape') setQuickCreateOpen(false);
                }}
                placeholder="Task title"
                className="w-full bg-transparent outline-none text-[15px] font-medium"
                style={{ color: 'var(--flow-text)' }}
              />
              <div className="flex items-center gap-2 flex-wrap">
                <ProjectSelect value={projectId} onChange={setProjectId} />
                <div className="flex items-center gap-1">
                  {priorities.map((p) => (
                    <button
                      key={p}
                      onClick={() => setPriority(p)}
                      className="h-7 w-7 rounded flex items-center justify-center"
                      style={{
                        background: priority === p ? 'var(--flow-hover)' : 'transparent',
                        border:
                          priority === p
                            ? '1px solid var(--flow-border)'
                            : '1px solid transparent',
                      }}
                    >
                      <PriorityIcon priority={p} />
                    </button>
                  ))}
                </div>
              </div>
            </div>
            <div
              className="flex items-center justify-between px-4 py-2.5"
              style={{ borderTop: '1px solid var(--flow-border)' }}
            >
              <span className="text-[10px]" style={{ color: 'var(--flow-text-dim)' }}>
                Press Enter to create
              </span>
              <button
                onClick={submit}
                disabled={!title.trim() || !projectId}
                className="px-3 py-1 rounded text-[11px] font-semibold disabled:opacity-40"
                style={{ background: '#64C5B9', color: '#0a0a0a' }}
              >
                Create
              </button>
            </div>
          </motion.div>
        </motion.div>
      )}
    </AnimatePresence>
  );
};

export default QuickCreate;
