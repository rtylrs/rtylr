import React, { useEffect, useRef, useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { motion, AnimatePresence } from 'motion/react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { ChevronRight, Calendar, User, CircleUser, FolderInput, Flag, Gauge } from 'lucide-react';
import { useFlowStore, Task } from '../stores/flowStore';
import { useUserStore } from '../stores/userStore';
import ProjectSelect from './ProjectSelect';
import PriorityIcon from './PriorityIcon';
import { cn, getInitials } from '../lib/utils';
import { getCurrentUserId } from '../lib/currentUser';
import { displayUserName } from '../lib/userDisplay';
import { EFFORT_PRESETS, formatEffortMinutes } from '../lib/effortPresets';
import { useCreateTask } from '../../features/tasks/useTasks';

function FieldRow({
  label,
  icon: Icon,
  children,
}: {
  label: string;
  icon?: React.ComponentType<{ size?: number; className?: string; style?: React.CSSProperties }>;
  children: React.ReactNode;
}) {
  return (
    <div
      className="flex items-start gap-3 py-2.5"
      style={{ borderBottom: '1px solid var(--flow-border)' }}
    >
      <div
        className="flex items-center gap-2 w-[100px] shrink-0 pt-0.5"
        style={{ color: 'var(--flow-text-muted)' }}
      >
        {Icon && <Icon size={14} className="shrink-0 opacity-70" />}
        <span className="text-[11px] font-medium">{label}</span>
      </div>
      <div className="flex-1 min-w-0 text-[13px]">{children}</div>
    </div>
  );
}

export const QuickCreate: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { quickCreateOpen, setQuickCreateOpen, projects, activeWorkspaceId, sections } =
    useFlowStore();
  const user = useUserStore((s) => s.user);
  const createTaskMutation = useCreateTask();

  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [projectId, setProjectId] = useState<string>('');
  const [priority, setPriority] = useState<Task['priority']>('medium');
  const [dueDate, setDueDate] = useState('');
  const [assignToMe, setAssignToMe] = useState(true);
  const [effortMinutes, setEffortMinutes] = useState<number | undefined>(30);

  const titleRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    if (quickCreateOpen) {
      setTitle('');
      setDescription('');
      const first = projects.find((p) => p.workspaceId === activeWorkspaceId) || projects[0];
      setProjectId(first?.id || '');
      setPriority('medium');
      setDueDate('');
      setAssignToMe(true);
      setEffortMinutes(30);
      setTimeout(() => titleRef.current?.focus(), 50);
      const prev = document.body.style.overflow;
      document.body.style.overflow = 'hidden';
      return () => {
        document.body.style.overflow = prev;
      };
    }
  }, [quickCreateOpen, projects, activeWorkspaceId]);

  const submit = async () => {
    if (!title.trim()) {
      toast.error(t('toasts.task.titleRequired'));
      return;
    }
    if (!projectId) return;
    const project = projects.find((p) => p.id === projectId);
    if (!project) return;

    const uid = getCurrentUserId();

    try {
      // Backend returns { task: ApiTask }. Unwrap before use - direct
      // access on the envelope would crash at runtime (TS catches it now).
      const res = await createTaskMutation.mutateAsync({
        title: title.trim(),
        workspaceId: project.workspaceId,
        projectId,
        priority,
        description: description.trim() || undefined,
        dueDate: dueDate || undefined,
        assigneeId: assignToMe ? (uid ?? undefined) : undefined,
        effortMinutes,
      });
      const task = res.task;
      const taskId = String(task.id);
      toast.success(t('toasts.task.createdWithId', { id: task.identifier ?? taskId }), {
        description: task.title,
      });
      setQuickCreateOpen(false);
      navigate({ to: `/task/${taskId}` as string });
    } catch {
      // error toast handled by mutation onError
    }
  };

  const priorities: Task['priority'][] = ['urgent', 'high', 'medium', 'low', 'no_priority'];
  const assigneeName = displayUserName(user);

  return (
    <AnimatePresence>
      {quickCreateOpen && (
        <motion.div
          key="quick-create-overlay"
          className="fixed inset-0 z-[100] flex overflow-hidden"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.15 }}
        >
          <motion.button
            type="button"
            aria-label="Close"
            className="flex-1 cursor-default border-0"
            style={{ background: 'rgba(0,0,0,0.45)' }}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={() => setQuickCreateOpen(false)}
          />
          <motion.aside
            role="dialog"
            aria-modal="true"
            aria-labelledby="quick-create-title"
            className="h-full w-full max-w-full flex flex-col shadow-2xl shrink-0 sm:max-w-[480px]"
            style={{
              background: 'var(--flow-surface)',
              borderLeft: '1px solid var(--flow-border)',
            }}
            initial={{ x: '100%' }}
            animate={{ x: 0 }}
            exit={{ x: '100%' }}
            transition={{ type: 'spring', damping: 32, stiffness: 380 }}
            onKeyDown={(e) => {
              if (e.key === 'Escape') setQuickCreateOpen(false);
            }}
          >
            {/* Header */}
            <div
              className="flex items-center justify-between gap-3 px-4 py-3 shrink-0"
              style={{ borderBottom: '1px solid var(--flow-border)' }}
            >
              <div className="min-w-0">
                <p
                  id="quick-create-title"
                  className="text-[13px] font-semibold"
                  style={{ color: 'var(--flow-text)' }}
                >
                  New task
                </p>
                <p className="text-[11px] mt-0.5" style={{ color: 'var(--flow-text-dim)' }}>
                  Add details, then create
                </p>
              </div>
              <button
                type="button"
                onClick={() => setQuickCreateOpen(false)}
                className="h-9 w-9 shrink-0 rounded-lg flex items-center justify-center transition-colors hover:bg-[var(--flow-hover)]"
                style={{ color: 'var(--flow-text-muted)' }}
                aria-label="Close"
              >
                <ChevronRight size={18} strokeWidth={2} />
              </button>
            </div>

            {/* Body */}
            <div className="flex-1 min-h-0 overflow-y-auto px-4 py-3">
              <textarea
                ref={titleRef}
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
                    e.preventDefault();
                    submit();
                  }
                }}
                placeholder={t('flow.quickCreate.titlePlaceholder')}
                rows={2}
                className="w-full bg-transparent outline-none resize-none text-[20px] font-semibold leading-snug placeholder:opacity-40 mb-1"
                style={{ color: 'var(--flow-text)' }}
              />

              <div className="mt-1">
                <FieldRow label="Assignee" icon={User}>
                  <div className="flex items-center gap-2 flex-wrap">
                    <button
                      type="button"
                      onClick={() => setAssignToMe(true)}
                      className={cn(
                        'inline-flex items-center gap-2 rounded-md px-2.5 py-1.5 text-[12px] font-medium transition-colors',
                        assignToMe ? '' : 'hover:bg-[var(--flow-hover)]'
                      )}
                      style={{
                        background: assignToMe ? 'var(--flow-hover)' : 'transparent',
                        color: 'var(--flow-text)',
                        border: assignToMe
                          ? `1px solid ${'var(--accent)'}44`
                          : '1px solid transparent',
                      }}
                    >
                      <span
                        className="h-6 w-6 rounded-full flex items-center justify-center text-[10px] font-bold"
                        style={{ background: 'var(--accent-muted)', color: 'var(--accent)' }}
                      >
                        {getInitials(assigneeName)}
                      </span>
                      {assigneeName}
                    </button>
                    <button
                      type="button"
                      onClick={() => setAssignToMe(false)}
                      className={cn(
                        'inline-flex items-center gap-1.5 rounded-md px-2.5 py-1.5 text-[12px] transition-colors',
                        !assignToMe ? '' : 'hover:bg-[var(--flow-hover)]'
                      )}
                      style={{
                        background: !assignToMe ? 'var(--flow-hover)' : 'transparent',
                        color: 'var(--flow-text-muted)',
                        border: !assignToMe
                          ? '1px solid var(--flow-border)'
                          : '1px solid transparent',
                      }}
                    >
                      <CircleUser size={14} />
                      Unassigned
                    </button>
                  </div>
                </FieldRow>

                <FieldRow label="Effort" icon={Gauge}>
                  <div className="flex flex-wrap gap-1">
                    <button
                      type="button"
                      onClick={() => setEffortMinutes(undefined)}
                      className="text-[11px] px-2 py-1 rounded-md"
                      style={{
                        background: effortMinutes == null ? 'var(--flow-hover)' : 'transparent',
                        color: 'var(--flow-text-muted)',
                        border: '1px solid var(--flow-border)',
                      }}
                    >
                      ,
                    </button>
                    {EFFORT_PRESETS.map((m) => (
                      <button
                        key={m}
                        type="button"
                        onClick={() => setEffortMinutes(m)}
                        className="text-[11px] px-2 py-1 rounded-md"
                        style={{
                          background: effortMinutes === m ? 'var(--accent-muted)' : 'transparent',
                          color: 'var(--flow-text)',
                          border:
                            effortMinutes === m
                              ? `1px solid ${'var(--accent)'}55`
                              : '1px solid var(--flow-border)',
                        }}
                      >
                        {formatEffortMinutes(m)}
                      </button>
                    ))}
                  </div>
                </FieldRow>

                <FieldRow label="Due date" icon={Calendar}>
                  <input
                    type="date"
                    value={dueDate}
                    onChange={(e) => setDueDate(e.target.value)}
                    className="w-full max-w-[200px] bg-transparent outline-none rounded-md px-2 py-1.5 text-[13px]"
                    style={{
                      border: '1px solid var(--flow-border)',
                      color: 'var(--flow-text)',
                    }}
                  />
                </FieldRow>

                <FieldRow label="Project" icon={FolderInput}>
                  <div className="-ml-1">
                    <ProjectSelect value={projectId} onChange={setProjectId} />
                  </div>
                </FieldRow>

                <FieldRow label="Priority" icon={Flag}>
                  <div className="flex items-center gap-1 flex-wrap">
                    {priorities.map((p) => (
                      <button
                        key={p}
                        type="button"
                        onClick={() => setPriority(p)}
                        className="h-8 w-8 rounded-md flex items-center justify-center transition-colors"
                        style={{
                          background: priority === p ? 'var(--flow-hover)' : 'transparent',
                          border:
                            priority === p
                              ? '1px solid var(--flow-border)'
                              : '1px solid transparent',
                        }}
                        title={p}
                      >
                        <PriorityIcon priority={p} />
                      </button>
                    ))}
                  </div>
                </FieldRow>
              </div>

              <div className="mt-4">
                <p
                  className="text-[11px] font-semibold uppercase tracking-wide mb-2"
                  style={{ color: 'var(--flow-text-muted)' }}
                >
                  Description
                </p>
                <textarea
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder={t('flow.quickCreate.descriptionPlaceholder')}
                  rows={6}
                  className="w-full rounded-lg px-3 py-2.5 text-[13px] leading-relaxed outline-none resize-y min-h-[120px] placeholder:opacity-45"
                  style={{
                    background: 'var(--flow-input-bg)',
                    border: '1px solid var(--flow-border)',
                    color: 'var(--flow-text)',
                  }}
                />
              </div>
            </div>

            {/* Footer */}
            <div
              className="flex items-center justify-end gap-2 px-4 py-3 shrink-0"
              style={{ borderTop: '1px solid var(--flow-border)' }}
            >
              <button
                type="button"
                onClick={() => setQuickCreateOpen(false)}
                className="px-3 py-2 rounded-lg text-[12px] font-medium transition-colors hover:bg-[var(--flow-hover)]"
                style={{ color: 'var(--flow-text-muted)' }}
              >
                Cancel
              </button>
              <button
                type="button"
                onClick={submit}
                disabled={!title.trim() || !projectId || createTaskMutation.isPending}
                className="px-4 py-2 rounded-lg text-[12px] font-semibold disabled:opacity-40 disabled:cursor-not-allowed transition-opacity"
                style={{ background: 'var(--accent)', color: '#0a0a0a' }}
              >
                {createTaskMutation.isPending ? 'Creating…' : 'Create task'}
              </button>
            </div>
          </motion.aside>
        </motion.div>
      )}
    </AnimatePresence>
  );
};

export default QuickCreate;
