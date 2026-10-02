import { EmptyState } from '../../../../shared/components/dash/EmptyState';
import React, { useMemo } from 'react';
import { Link, useNavigate } from '@tanstack/react-router';
import { motion } from 'motion/react';
import { Sun, Pin } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';

import { useFlowStore, Task, Project } from '../../shared/stores/flowStore';
import { useMyIssues } from '../tasks/useMyIssues';
import { apiTaskToTask } from '../../shared/lib/taskMapper';
import { isDueToday } from '../../shared/lib/lifeRhythm';
import { isTaskMine } from '../../shared/lib/currentUser';
import { formatEffortMinutes } from '../../shared/lib/effortPresets';
import { formatDueDate } from '../../shared/lib/utils';
import PriorityIcon from '../../shared/components/PriorityIcon';
import { ListLoadFailed } from '../../shared/components/ListLoadFailed';
import { hasQueryFailed } from '../../shared/lib/queryFailureUtils';

const ACCENT_VAR = 'var(--accent)';

/**
 * One titled group with its count. Two of these replace a single flat list
 * that called everything "pinned" while also containing due-today tasks.
 *
 * Renders nothing when its group is empty, so the screen shows only the
 * reasons that actually apply rather than an empty heading.
 */
const FocusSection: React.FC<{
  titleKey: string;
  tasks: Task[];
  projects: Project[];
  startIndex: number;
}> = ({ titleKey, tasks, projects, startIndex }) => {
  const { t } = useTranslation();
  if (tasks.length === 0) return null;
  return (
    <section>
      <h2
        className="flex items-baseline gap-2 px-4 py-2 text-flow-meta font-semibold uppercase tracking-wider"
        style={{ color: 'var(--flow-text-muted)' }}
      >
        {t(titleKey)}
        <span className="font-normal tabular-nums" style={{ color: 'var(--flow-text-dim)' }}>
          {tasks.length}
        </span>
      </h2>
      <ul className="divide-y" style={{ borderColor: 'var(--flow-border)' }}>
        {tasks.map((task, i) => (
          <FocusRow
            key={task.id}
            task={task}
            project={projects.find((p) => p.id === task.projectId)}
            index={startIndex + i}
          />
        ))}
      </ul>
    </section>
  );
};

const FocusToday: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { projects } = useFlowStore();
  const myIssuesQuery = useMyIssues();
  const { data: myTasks, isLoading } = myIssuesQuery;
  // A failed load must not render as "nothing due today": that tells the user
  // their day is clear when we simply could not read it.
  const loadFailed = hasQueryFailed(myIssuesQuery);
  const tasks: Task[] = (myTasks ?? []).map(apiTaskToTask);

  const focusTasks = useMemo(() => {
    // No workspace filter. `t.workspaceId` is undefined because the API no
    // longer sends a workspace field, and `activeWorkspaceId` is '' because
    // that layer was removed, so `undefined === ''` was false for every task
    // and this list could never contain anything.
    const list = tasks.filter(
      (t) =>
        !t.archived &&
        t.status !== 'done' &&
        t.status !== 'cancelled' &&
        isTaskMine(t.assigneeId) &&
        (t.focusPin === true || isDueToday(t.dueDate))
    );
    return [...list].sort((a, b) => {
      const ap = a.focusPin ? 0 : 1;
      const bp = b.focusPin ? 0 : 1;
      if (ap !== bp) return ap - bp;
      const ad = a.dueDate || '';
      const bd = b.dueDate || '';
      return ad.localeCompare(bd);
    });
  }, [tasks]);

  // S-05: pinned and due-today are different reasons for a task to be here,
  // and the screen promised both while calling everything "pinned". A task
  // that is pinned and also due today belongs to Pinned only, so nothing is
  // counted or listed twice.
  const pinnedTasks = useMemo(
    () => focusTasks.filter((task) => task.focusPin === true),
    [focusTasks]
  );
  const dueTodayTasks = useMemo(
    () => focusTasks.filter((task) => task.focusPin !== true),
    [focusTasks]
  );

  const backlogHint = useMemo(() => {
    return tasks.filter((t) => !t.archived && t.status === 'backlog' && isTaskMine(t.assigneeId))
      .length;
  }, [tasks]);

  return (
    <div className="flex flex-col h-full max-w-3xl mx-auto px-4 py-6 sm:px-6 sm:py-8">
      <HelmetContainer title={`${t('flowNav.focusToday')} · Flow`} />
      <motion.div initial={{ opacity: 0, y: -6 }} animate={{ opacity: 1, y: 0 }} className="mb-8">
        <div className="flex items-center gap-2 mb-1">
          {/* G-C2: the accent is for actions and focus, not decoration. */}
          <Sun size={22} style={{ color: 'var(--flow-text-muted)' }} strokeWidth={2} />
          <h1 className="text-flow-screen font-bold" style={{ color: 'var(--flow-text)' }}>
            {t('flowNav.focusToday', 'Today')}
          </h1>
        </div>
        {/* G-T3 caps prose at 75 characters; this ran to about 110 per line. */}
        <p
          className="max-w-[68ch] text-flow-body leading-relaxed"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          {t('flowCommon.focusToday.subtitle')}
        </p>
        {backlogHint > 0 && (
          <p className="text-flow-meta mt-3" style={{ color: 'var(--flow-text-dim)' }}>
            {t('flowCommon.focusToday.backlogHint', { count: backlogHint })}{' '}
            <Link
              to="/flow/triage"
              className="underline-offset-2 hover:underline"
              style={{ color: ACCENT_VAR }}
            >
              {t('flow.shell.triage')}
            </Link>{' '}
            {t('flowCommon.focusToday.backlogHintSuffix')}
          </p>
        )}
      </motion.div>

      <div className="flex-1 min-h-0 overflow-y-auto">
        {loadFailed ? (
          <ListLoadFailed onRetry={() => void myIssuesQuery.refetch()} />
        ) : isLoading ? null : focusTasks.length === 0 ? (
          // G-E1: sized to its content on the page background. It used to sit
          // in a bordered card stretched to the full viewport height, with
          // roughly 250px of blank card above the illustration and 400px
          // below it.
          <EmptyState
            icon={Sun}
            title={t('flowCommon.focusToday.emptyTitle')}
            description={t('flowCommon.focusToday.emptyHint')}
            action={{
              label: t('flow.shell.myTasks'),
              onClick: () => navigate({ to: '/flow/my-issues' }),
            }}
            size="compact"
          />
        ) : (
          <>
            <FocusSection
              titleKey="flowCommon.focusToday.pinnedHeading"
              tasks={pinnedTasks}
              projects={projects}
              startIndex={0}
            />
            <FocusSection
              titleKey="flowCommon.focusToday.dueTodayHeading"
              tasks={dueTodayTasks}
              projects={projects}
              startIndex={pinnedTasks.length}
            />
          </>
        )}
      </div>
    </div>
  );
};

const FocusRow: React.FC<{
  task: Task;
  project: Project | undefined;
  index: number;
}> = ({ task, project, index }) => {
  const { t } = useTranslation();
  return (
    <motion.li
      initial={{ opacity: 0, x: -6 }}
      animate={{ opacity: 1, x: 0 }}
      transition={{ delay: index * 0.04 }}
    >
      <Link
        to={`/flow/task/${task.id}` as string}
        className="flex items-center gap-3 px-4 py-3.5 hover:bg-[var(--flow-hover)] transition-colors"
      >
        <PriorityIcon priority={task.priority} />
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            {task.focusPin && (
              <Pin size={12} className="shrink-0" style={{ color: ACCENT_VAR }} aria-hidden />
            )}
            <span
              className="text-flow-body font-medium truncate"
              style={{ color: 'var(--flow-text)' }}
            >
              {task.title}
            </span>
          </div>
          <div
            className="flex items-center gap-2 mt-1 text-flow-meta"
            style={{ color: 'var(--flow-text-dim)' }}
          >
            <span className="truncate">{project?.name ?? task.projectName}</span>
            {task.dueDate && (
              <>
                <span>·</span>
                <span>
                  {isDueToday(task.dueDate)
                    ? t('flowCommon.focusToday.dueToday')
                    : formatDueDate(task.dueDate, t)}
                </span>
              </>
            )}
            {task.effortMinutes != null && task.effortMinutes > 0 && (
              <>
                <span>·</span>
                <span>{formatEffortMinutes(task.effortMinutes)}</span>
              </>
            )}
          </div>
        </div>
      </Link>
    </motion.li>
  );
};

export default FocusToday;
