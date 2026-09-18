import React, { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { motion } from 'motion/react';
import { ChevronDown, ChevronRight } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { useMyIssues } from './useMyIssues';
import { useFlowStore } from '../../shared/stores/flowStore';
import { apiTaskToTask } from '../../shared/lib/taskMapper';
import { TaskRow } from '../../shared/components/TaskRow';
import {
  bucketForMyTask,
  MY_TASK_BUCKET_LABEL_KEY,
  MY_TASK_BUCKET_ORDER,
  MY_TASK_BUCKET_COLLAPSED_BY_DEFAULT,
  type MyTaskBucket,
} from '../../shared/lib/myTasksBuckets';
import { hasQueryFailed } from '../../shared/lib/queryFailureUtils';
import { ListLoadFailed } from '../../shared/components/ListLoadFailed';

const MyIssues: React.FC = () => {
  const { t } = useTranslation();
  const myIssuesQuery = useMyIssues();
  const { data: rawTasks = [], isLoading } = myIssuesQuery;
  const loadFailed = hasQueryFailed(myIssuesQuery);
  // GetMyTasks doesn't filter archived server-side, so exclude client-side
  // to match Triage/FocusToday/FlowLayout.
  const tasks = useMemo(() => rawTasks.map(apiTaskToTask).filter((t) => !t.archived), [rawTasks]);
  const { selectedTaskIds, setQuickCreateOpen } = useFlowStore();
  const [collapsed, setCollapsed] = useState<ReadonlySet<MyTaskBucket>>(
    MY_TASK_BUCKET_COLLAPSED_BY_DEFAULT
  );

  const toggle = (bucket: MyTaskBucket): void => {
    setCollapsed((current) => {
      const next = new Set(current);
      if (next.has(bucket)) next.delete(bucket);
      else next.add(bucket);
      return next;
    });
  };

  // Grouped by time bucket, not by status. `myTasksBuckets` has implemented
  // exactly this since it was written and nothing imported it, so the screen
  // rendered a status kanban and scattered overdue work across four columns.
  // Done and cancelled work is not a bucket of its own here; it is finished.
  const byBucket = useMemo(() => {
    const map = new Map<string, typeof tasks>(MY_TASK_BUCKET_ORDER.map((b) => [b, []]));
    for (const task of tasks) {
      if (task.status === 'done' || task.status === 'cancelled') continue;
      map.get(bucketForMyTask(task))?.push(task);
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
          <h1 className="text-title font-semibold" style={{ color: 'var(--flow-text)' }}>
            {t('flow.shell.myTasks')}
          </h1>
          <p className="text-meta" style={{ color: 'var(--flow-text-dim)' }}>
            {isLoading || loadFailed
              ? t('flow.myIssues.loading')
              : t('flow.myIssues.taskCount', { count: tasks.length })}
          </p>
        </div>
      </div>

      <div className="relative flex-1 overflow-y-auto">
        <div className="p-4 space-y-5">
          {MY_TASK_BUCKET_ORDER.map((bucket, bucketIdx) => {
            const bucketTasks = byBucket.get(bucket) ?? [];
            if (bucketTasks.length === 0) return null;
            const isCollapsed = collapsed.has(bucket);
            return (
              <motion.section
                key={bucket}
                initial={{ opacity: 0, y: 8 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.15, delay: bucketIdx * 0.04 }}
              >
                <h2>
                  <button
                    type="button"
                    onClick={() => toggle(bucket)}
                    aria-expanded={!isCollapsed}
                    aria-controls={`my-tasks-${bucket}`}
                    className="flow-focus flex min-h-11 w-full items-center gap-2 px-1 text-start text-meta font-semibold"
                    style={{ color: 'var(--flow-text)' }}
                  >
                    {isCollapsed ? <ChevronRight size={14} /> : <ChevronDown size={14} />}
                    {t(MY_TASK_BUCKET_LABEL_KEY[bucket])}
                    <span
                      className="rounded-sm px-1.5 py-0.5 text-meta font-medium"
                      style={{ background: 'var(--flow-hover)', color: 'var(--flow-text-muted)' }}
                    >
                      {bucketTasks.length}
                    </span>
                  </button>
                </h2>
                <div
                  id={`my-tasks-${bucket}`}
                  hidden={isCollapsed}
                  className="rounded-xl overflow-hidden"
                  style={{
                    background: 'var(--flow-card)',
                    border: '1px solid var(--flow-border)',
                  }}
                >
                  {bucketTasks.map((task) => (
                    <TaskRow
                      key={task.id}
                      task={task}
                      showProject
                      selected={selectedTaskIds.includes(task.id)}
                    />
                  ))}
                </div>
              </motion.section>
            );
          })}
        </div>

        {loadFailed && !isLoading && (
          <div className="absolute inset-0 z-10 flex items-center justify-center p-4">
            <div className="w-full max-w-md">
              <ListLoadFailed onRetry={() => void myIssuesQuery.refetch()} />
            </div>
          </div>
        )}

        {tasks.length === 0 && !isLoading && !loadFailed && (
          <div className="pointer-events-none absolute inset-0 z-10 flex items-center justify-center p-4">
            <div
              className="pointer-events-auto w-full max-w-md rounded-xl px-6 py-7 text-center shadow-xl"
              style={{
                background: 'var(--flow-surface)',
                border: '1px solid var(--accent-muted)',
              }}
            >
              <h2 className="text-title font-semibold" style={{ color: 'var(--flow-text)' }}>
                {t('flow.myIssues.emptyTitle')}
              </h2>
              <p className="mt-2 text-meta" style={{ color: 'var(--flow-text-muted)' }}>
                {t('flow.myIssues.emptyDescription')}
              </p>
              <button
                type="button"
                onClick={() => setQuickCreateOpen(true)}
                className="mt-5 inline-flex min-h-11 items-center justify-center rounded-lg px-5 text-meta font-semibold transition-opacity hover:opacity-90"
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
