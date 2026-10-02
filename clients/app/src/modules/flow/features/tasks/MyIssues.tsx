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
import { useUserStore } from '../../shared/stores/userStore';
import { canFlowPermission } from '../../shared/lib/roleUtils';
import { Button } from '../../../../shared/components/ui/button';
import { Plus, ListTodo } from 'lucide-react';

const MyIssues: React.FC = () => {
  const { t } = useTranslation();
  const access = useUserStore((state) => state.user?.effectiveAccess);
  const canCreate = canFlowPermission('flow:tasks:create', access);
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
        className="portal-flow-page-header flex items-center justify-between gap-4 px-6 py-5 shrink-0"
        style={{ borderBottom: '1px solid var(--flow-border)' }}
      >
        <div>
          <h1 className="text-flow-title font-semibold" style={{ color: 'var(--flow-text)' }}>
            {t('flow.shell.myTasks')}
          </h1>
          {!loadFailed && (
            <p className="text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
              {isLoading
                ? t('flow.myIssues.loading')
                : t('flow.myIssues.taskCount', { count: tasks.length })}
            </p>
          )}
        </div>
        {canCreate && (
          <Button onClick={() => setQuickCreateOpen(true)}>
            <Plus size={16} />
            {t('flow.shell.newTask')}
          </Button>
        )}
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
                    className="flow-focus flex min-h-11 w-full items-center gap-2 px-1 text-start text-flow-meta font-semibold"
                    style={{ color: 'var(--flow-text)' }}
                  >
                    {isCollapsed ? <ChevronRight size={14} /> : <ChevronDown size={14} />}
                    {t(MY_TASK_BUCKET_LABEL_KEY[bucket])}
                    <span
                      className="rounded-sm px-1.5 py-0.5 text-flow-meta font-medium"
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
              className="portal-flow-empty pointer-events-auto w-full max-w-md px-6 py-7 text-center"
              style={{
                background: 'var(--flow-surface)',
              }}
            >
              <div className="portal-flow-empty-icon">
                <ListTodo size={24} />
              </div>
              <h2 className="text-flow-title font-semibold" style={{ color: 'var(--flow-text)' }}>
                {t('flow.myIssues.emptyTitle')}
              </h2>
              <p className="mt-2 text-flow-meta" style={{ color: 'var(--flow-text-muted)' }}>
                {t('flow.myIssues.emptyDescription')}
              </p>
              {canCreate && (
                <Button type="button" onClick={() => setQuickCreateOpen(true)} className="mt-5">
                  <Plus size={16} />
                  {t('flow.board.addTaskLabel')}
                </Button>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default MyIssues;
