import React, { useMemo, useRef, useState } from 'react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { motion } from 'motion/react';
import { ListTodo, Plus } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { useMyIssues } from './useMyIssues';
import { useCreateTask } from './useTasks';
import { useMyTasksProject } from './useMyTasksProject';
import { useLabels } from './useLabels';
import { useFlowStore } from '../../shared/stores/flowStore';
import { getCurrentUserId } from '../../shared/lib/currentUser';
import { apiTaskToTask } from '../../shared/lib/taskMapper';
import TaskCard from '../../shared/components/TaskCard';

// Backlog lives here as the leftmost column instead of as its own
// sidebar tab - keeping unscheduled work next to active work means a
// single view covers personal task management end-to-end.
const KANBAN_COLUMNS: { status: string; labelKey: string; color: string }[] = [
  { status: 'backlog', labelKey: 'backlog', color: '#6b7280' },
  { status: 'todo', labelKey: 'todo', color: '#9ca3af' },
  { status: 'in_progress', labelKey: 'inProgress', color: '#3b82f6' },
  { status: 'in_review', labelKey: 'inReview', color: '#f59e0b' },
  { status: 'done', labelKey: 'done', color: '#22c55e' },
];

const MyIssues: React.FC = () => {
  const { t } = useTranslation();
  const { data: rawTasks = [], isLoading } = useMyIssues();
  // GetMyTasks doesn't filter archived server-side, so exclude client-side
  // to match Triage/FocusToday/FlowLayout.
  const tasks = useMemo(() => rawTasks.map(apiTaskToTask).filter((t) => !t.archived), [rawTasks]);
  const { selectedTaskIds } = useFlowStore();
  const quickAddInputRef = useRef<HTMLInputElement>(null);

  const byStatus = useMemo(() => {
    const map: Record<string, typeof tasks> = {};
    for (const col of KANBAN_COLUMNS) map[col.status] = [];
    for (const t of tasks) {
      if (map[t.status] !== undefined) map[t.status].push(t);
    }
    return map;
  }, [tasks]);

  return (
    <div className="flex flex-col h-full">
      <HelmetContainer title={`${t('flow.shell.myTasks')} · Flow`} />
      <div
        className="flex items-center justify-between px-6 py-4 shrink-0"
        style={{ borderBottom: '1px solid var(--flow-border)' }}
      >
        <div>
          <h1 className="text-[15px] font-semibold" style={{ color: 'var(--flow-text)' }}>
            {t('flow.shell.myTasks')}
          </h1>
          <p className="text-[11px]" style={{ color: 'var(--flow-text-dim)' }}>
            {isLoading
              ? t('flow.myIssues.loading')
              : t('flow.myIssues.taskCount', { count: tasks.length })}
          </p>
        </div>
      </div>

      <QuickAddRow inputRef={quickAddInputRef} />

      {tasks.length === 0 && !isLoading ? (
        <EmptyState
          icon={ListTodo}
          title={t('flow.myIssues.emptyTitle')}
          description={t('flow.myIssues.emptyDescription')}
          action={{
            label: t('flow.myIssues.emptyCta'),
            onClick: () => quickAddInputRef.current?.focus(),
          }}
        />
      ) : (
        <div className="flex-1 overflow-x-auto overflow-y-hidden">
          <div className="flex gap-3 h-full p-4 min-w-max">
            {KANBAN_COLUMNS.map((col, colIdx) => {
              const colTasks = byStatus[col.status] ?? [];
              return (
                <motion.div
                  key={col.status}
                  initial={{ opacity: 0, y: 8 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.15, delay: colIdx * 0.04 }}
                  className="flex flex-col rounded-xl shrink-0"
                  style={{
                    width: 280,
                    background: 'var(--flow-card)',
                    border: '1px solid var(--flow-border)',
                  }}
                >
                  {/* Column header */}
                  <div
                    className="flex items-center gap-2 px-3 py-2.5 shrink-0"
                    style={{ borderBottom: '1px solid var(--flow-border)' }}
                  >
                    <span
                      className="h-2 w-2 rounded-full shrink-0"
                      style={{ background: col.color }}
                    />
                    <span
                      className="text-[12px] font-semibold flex-1"
                      style={{ color: 'var(--flow-text)' }}
                    >
                      {t(`flow.taskDetail.${col.labelKey}`)}
                    </span>
                    <span
                      className="text-[10px] px-1.5 py-0.5 rounded font-medium"
                      style={{
                        background: `${col.color}1a`,
                        color: col.color,
                      }}
                    >
                      {colTasks.length}
                    </span>
                  </div>

                  {/* Tasks */}
                  <div className="flex-1 overflow-y-auto p-2 space-y-1.5">
                    {colTasks.length === 0 ? (
                      <p
                        className="text-center text-[11px] py-6"
                        style={{ color: 'var(--flow-text-dim)' }}
                      >
                        {t('flow.myIssues.noTasks')}
                      </p>
                    ) : (
                      colTasks.map((t) => (
                        <TaskCard
                          key={t.id}
                          task={t}
                          compact
                          showProject
                          selected={selectedTaskIds.includes(t.id)}
                        />
                      ))
                    )}
                  </div>
                </motion.div>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
};

interface QuickAddRowProps {
  inputRef: React.RefObject<HTMLInputElement | null>;
}

const QuickAddRow: React.FC<QuickAddRowProps> = ({ inputRef }) => {
  const { t } = useTranslation();
  const [title, setTitle] = useState('');
  const createTask = useCreateTask();
  const { activeWorkspaceId } = useFlowStore();
  const { data: myTasksProject } = useMyTasksProject();
  const myTasksWorkspaceId = myTasksProject?.workspaceId
    ? String(myTasksProject.workspaceId)
    : activeWorkspaceId;
  const { data: labels = [], isLoading: labelsLoading } = useLabels(
    myTasksWorkspaceId || undefined
  );
  const hasNoLabels = !labelsLoading && labels.length === 0;

  const submit = (): void => {
    const trimmed = title.trim();
    if (!trimmed) {
      toast.error(t('toasts.task.titleRequired'));
      return;
    }
    if (hasNoLabels) {
      toast.error(t('flow.myIssues.quickAddNoLabelsReason'));
      return;
    }
    if (!myTasksProject || labels.length === 0) {
      toast.error(t('errors.generic'));
      return;
    }
    // Personal tasks land in Backlog so the leftmost column actually
    // collects new captures instead of jamming them into Todo.
    createTask.mutate(
      {
        workspaceId: String(myTasksProject.workspaceId),
        projectId: String(myTasksProject.projectId),
        title: trimmed,
        status: 'backlog',
        assigneeId: getCurrentUserId() ?? undefined,
        labels: [String(labels[0].id)],
      },
      { onSuccess: () => setTitle('') }
    );
  };

  return (
    <div
      className="flex items-center gap-2 px-6 py-2 shrink-0"
      style={{ borderBottom: '1px solid var(--flow-border)' }}
    >
      <Plus size={12} style={{ color: 'var(--flow-text-dim)' }} />
      <input
        ref={inputRef}
        value={title}
        onChange={(e) => setTitle(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === 'Enter') {
            e.preventDefault();
            submit();
          }
        }}
        placeholder={t('flow.myIssues.quickAddPlaceholder')}
        disabled={createTask.isPending}
        className="flex-1 bg-transparent outline-none text-[13px] placeholder:opacity-50"
        style={{ color: 'var(--flow-text)' }}
      />
      {title.trim().length > 0 && (
        <button
          type="button"
          onClick={submit}
          disabled={createTask.isPending || hasNoLabels}
          title={hasNoLabels ? t('flow.myIssues.quickAddNoLabelsReason') : undefined}
          className="text-[11px] font-semibold px-2.5 py-1 rounded transition-opacity hover:opacity-90 disabled:opacity-50"
          style={{ background: 'var(--accent)', color: 'var(--flow-on-accent)' }}
        >
          {t('flow.myIssues.quickAddButton')}
        </button>
      )}
    </div>
  );
};

export default MyIssues;
