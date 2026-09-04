import React, { useMemo } from 'react';
import { Link, useNavigate } from '@tanstack/react-router';
import { motion } from 'motion/react';
import { Sun, Pin } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { useFlowStore, Task, Project } from '../../shared/stores/flowStore';
import { useMyIssues } from '../tasks/useMyIssues';
import { apiTaskToTask } from '../../shared/lib/taskMapper';
import { isDueToday } from '../../shared/lib/lifeRhythm';
import { isTaskMine } from '../../shared/lib/currentUser';
import { formatEffortMinutes } from '../../shared/lib/effortPresets';
import { formatDate } from '../../shared/lib/utils';
import PriorityIcon from '../../shared/components/PriorityIcon';

const ACCENT_VAR = 'var(--accent)';

const FocusToday: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { projects } = useFlowStore();
  const { data: myTasks } = useMyIssues();
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

  const backlogHint = useMemo(() => {
    return tasks.filter((t) => !t.archived && t.status === 'backlog' && isTaskMine(t.assigneeId))
      .length;
  }, [tasks]);

  return (
    <div className="flex flex-col h-full max-w-3xl mx-auto px-4 py-6 sm:px-6 sm:py-8">
      <HelmetContainer title={`${t('nav.focusToday')} · Flow`} />
      <motion.div initial={{ opacity: 0, y: -6 }} animate={{ opacity: 1, y: 0 }} className="mb-8">
        <div className="flex items-center gap-2 mb-1">
          <Sun size={22} style={{ color: ACCENT_VAR }} strokeWidth={2} />
          <h1 className="text-xl font-bold" style={{ color: 'var(--flow-text)' }}>
            {t('nav.focusToday', 'Today')}
          </h1>
        </div>
        <p className="text-body leading-relaxed" style={{ color: 'var(--flow-text-muted)' }}>
          {t('common.focusToday.subtitle')}
        </p>
        {backlogHint > 0 && (
          <p className="text-meta mt-3" style={{ color: 'var(--flow-text-dim)' }}>
            {t('common.focusToday.backlogHint', { count: backlogHint })}{' '}
            <Link
              to="/triage"
              className="underline-offset-2 hover:underline"
              style={{ color: ACCENT_VAR }}
            >
              {t('flow.shell.triage')}
            </Link>{' '}
            {t('common.focusToday.backlogHintSuffix')}
          </p>
        )}
      </motion.div>

      <div
        className="flex-1 min-h-0 overflow-y-auto rounded-xl border"
        style={{ borderColor: 'var(--flow-border)', background: 'var(--flow-card)' }}
      >
        {focusTasks.length === 0 ? (
          <EmptyState
            icon={Sun}
            title={t('common.focusToday.emptyTitle')}
            description={t('common.focusToday.emptyHint')}
            action={{
              label: t('flow.shell.myTasks'),
              onClick: () => navigate({ to: '/my-issues' }),
            }}
          />
        ) : (
          <ul className="divide-y" style={{ borderColor: 'var(--flow-border)' }}>
            {focusTasks.map((t, i) => (
              <FocusRow
                key={t.id}
                task={t}
                project={projects.find((p) => p.id === t.projectId)}
                index={i}
              />
            ))}
          </ul>
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
        to={`/task/${task.id}` as string}
        className="flex items-center gap-3 px-4 py-3.5 hover:bg-[var(--flow-hover)] transition-colors"
      >
        <PriorityIcon priority={task.priority} />
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            {task.focusPin && (
              <Pin size={12} className="shrink-0" style={{ color: ACCENT_VAR }} aria-hidden />
            )}
            <span className="text-body font-medium truncate" style={{ color: 'var(--flow-text)' }}>
              {task.title}
            </span>
          </div>
          <div
            className="flex items-center gap-2 mt-1 text-meta"
            style={{ color: 'var(--flow-text-dim)' }}
          >
            <span className="truncate">{project?.name ?? task.projectName}</span>
            {task.dueDate && (
              <>
                <span>·</span>
                <span>
                  {isDueToday(task.dueDate)
                    ? t('common.focusToday.dueToday')
                    : formatDate(task.dueDate)}
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
