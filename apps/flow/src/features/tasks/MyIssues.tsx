import React, { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { motion } from 'motion/react';
import { HelmetContainer } from '@voxire/components';
import { useMyIssues } from './useMyIssues';
import { useFlowStore } from '../../shared/stores/flowStore';
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
  const { selectedTaskIds, setQuickCreateOpen } = useFlowStore();

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

      <div className="relative flex-1 overflow-x-auto overflow-y-hidden">
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
                    className="text-[10px] px-1.5 py-0.5 rounded-sm font-medium"
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

        {tasks.length === 0 && !isLoading && (
          <div className="pointer-events-none absolute inset-0 z-10 flex items-center justify-center p-4">
            <div
              className="pointer-events-auto w-full max-w-md rounded-xl px-6 py-7 text-center shadow-xl"
              style={{
                background: 'var(--flow-surface)',
                border: '1px solid var(--accent-muted)',
              }}
            >
              <h2 className="text-[16px] font-semibold" style={{ color: 'var(--flow-text)' }}>
                {t('flow.myIssues.emptyTitle')}
              </h2>
              <p className="mt-2 text-[12px]" style={{ color: 'var(--flow-text-muted)' }}>
                {t('flow.myIssues.emptyDescription')}
              </p>
              <button
                type="button"
                onClick={() => setQuickCreateOpen(true)}
                className="mt-5 inline-flex min-h-11 items-center justify-center rounded-lg px-5 text-[12px] font-semibold transition-opacity hover:opacity-90"
                style={{ background: 'var(--accent)', color: 'var(--flow-on-accent)' }}
              >
                {t('flow.board.addTaskLabel')}
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default MyIssues;
