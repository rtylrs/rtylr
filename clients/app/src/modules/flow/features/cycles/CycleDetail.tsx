import React, { useMemo, useState } from 'react';
import { Link, useNavigate, useParams } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { ArrowLeft, ArrowRight, Pencil, Trash2, Timer } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { useCycle, useCycles, useDeleteCycle, useUpdateCycle } from './useCycles';
import { useTasks } from '../tasks/useTasks';
import { useProjectStatuses } from '../projects/useStatuses';
import { taskBucket, orderByPlanRank } from '../projects/backlog/backlogModel';
import { apiTaskToTask } from '../../shared/lib/taskMapper';
import { useCanAdministerProject, useCanEditTask } from '../../shared/lib/currentUser';
import { useProjectRealtime } from '../../shared/lib/useProjectRealtime';
import { formatShortDate } from '../../shared/lib/utils';
import { todayIso } from '../../shared/lib/dateUtils';
import { hasQueryFailed } from '../../shared/lib/queryFailureUtils';
import { ListLoadFailed } from '../../shared/components/ListLoadFailed';
import TaskRow from '../../shared/components/TaskRow';
import ConfirmDialog from '../../shared/components/ConfirmDialog';
import type { ApiStatus } from '../../shared/services/flowApi';
import BurndownChart from '../../shared/components/BurndownChart';
import CycleBadge from '../../shared/components/CycleBadge';
import { EditCycleDialog } from '../projects/backlog/CycleDialogs';
import { buildBurndown } from './cycleModel';
import { CycleProgress } from './CycleProgress';

/**
 * One cycle: its tasks, progress, a burndown and what was carried over into it
 * (T-336). Planning itself (dragging tasks in and out, start, complete) stays
 * on the project's Backlog, which this page links to.
 */
const CycleDetail: React.FC = () => {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const { cycleId = '' } = useParams({ strict: false });
  const cycleQuery = useCycle(cycleId);
  const cycle = cycleQuery.data;
  const projectId = cycle?.projectId ?? '';
  useProjectRealtime(projectId);
  const tasksQuery = useTasks({ projectId: projectId || undefined });
  const { data: projectCycles = [] } = useCycles(projectId || undefined);
  const { data: rawStatuses } = useProjectStatuses(projectId);
  const canEdit = useCanEditTask(projectId);
  const canDelete = useCanAdministerProject(projectId);
  const updateCycle = useUpdateCycle();
  const deleteCycle = useDeleteCycle();
  const [editing, setEditing] = useState(false);
  const [confirmingDelete, setConfirmingDelete] = useState(false);

  const statusesById = useMemo(
    () => new Map<string, ApiStatus>((rawStatuses ?? []).map((s) => [s.id, s])),
    [rawStatuses]
  );
  // Every task in the cycle: deleting it sends them all back to the backlog.
  const members = useMemo(
    () =>
      orderByPlanRank(
        (tasksQuery.data ?? []).filter((x) => x.cycleId === cycleId).map(apiTaskToTask)
      ),
    [tasksQuery.data, cycleId]
  );
  // The cycle's scope, counted as the server counts it: not archived, not
  // cancelled.
  const scoped = useMemo(
    () => members.filter((x) => !x.archived && taskBucket(x, statusesById) !== 'cancelled'),
    [members, statusesById]
  );
  const open = scoped.filter((x) => taskBucket(x, statusesById) !== 'done');
  const finished = scoped.filter((x) => taskBucket(x, statusesById) === 'done');
  const burndown = useMemo(
    () =>
      cycle
        ? buildBurndown(
            cycle.startDate,
            cycle.endDate,
            scoped.map((x) => ({
              weight: 1,
              completedAt: taskBucket(x, statusesById) === 'done' ? x.completedAt : undefined,
            })),
            todayIso()
          )
        : [],
    [cycle, scoped, statusesById]
  );
  // The soonest upcoming cycle after this one; the list is not in date order.
  const nextCycle = projectCycles
    .filter((c) => c.status === 'upcoming' && cycle !== undefined && c.startDate >= cycle.endDate)
    .sort((a, b) => a.startDate.localeCompare(b.startDate))[0];

  if (cycleQuery.isPending) {
    return (
      <div className="p-6 text-sm" style={{ color: 'var(--flow-text-dim)' }}>
        {t('flow.cycles.loading')}
      </div>
    );
  }
  if (hasQueryFailed(cycleQuery) || !cycle) {
    return <ListLoadFailed onRetry={() => void cycleQuery.refetch()} />;
  }

  const carriedOver = cycle.carriedOver ?? [];

  return (
    <div className="flex flex-col h-full overflow-auto">
      <HelmetContainer title={`${cycle.name} · ${t('flow.shell.cycles')} · Flow`} />
      <div
        className="px-6 py-4 shrink-0 space-y-2"
        style={{ borderBottom: '1px solid var(--flow-border)' }}
      >
        <Link
          to="/flow/cycles"
          className="inline-flex items-center gap-1 text-flow-meta"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          <ArrowLeft size={12} className="rtl:rotate-180" /> {t('flow.cycles.detail.back')}
        </Link>
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="flex items-center gap-2 min-w-0">
            <Timer size={16} style={{ color: 'var(--flow-text-muted)' }} />
            <h1
              className="text-flow-title font-semibold truncate"
              style={{ color: 'var(--flow-text)' }}
            >
              {cycle.name}
            </h1>
            <CycleBadge cycle={cycle} compact showState />
          </div>
          <div className="flex items-center gap-2">
            {projectId && (
              <Link
                to="/flow/p/$projectId/backlog"
                params={{ projectId }}
                className="px-3 py-1.5 rounded-sm text-flow-meta font-semibold"
                style={{
                  border: '1px solid var(--flow-border-control)',
                  color: 'var(--flow-text)',
                }}
              >
                {t('flow.cycles.detail.openBacklog')}
              </Link>
            )}
            {canEdit && (
              <button
                type="button"
                onClick={() => setEditing(true)}
                className="flex items-center gap-1.5 px-3 py-1.5 rounded-sm text-flow-meta"
                style={{
                  border: '1px solid var(--flow-border-control)',
                  color: 'var(--flow-text)',
                }}
              >
                <Pencil size={11} /> {t('flow.cycles.detail.edit')}
              </button>
            )}
            {canDelete && (
              <button
                type="button"
                onClick={() => setConfirmingDelete(true)}
                className="flex items-center gap-1.5 px-3 py-1.5 rounded-sm text-flow-meta"
                style={{
                  border: '1px solid var(--flow-border-control)',
                  color: 'var(--flow-danger-fg)',
                }}
              >
                <Trash2 size={11} /> {t('flow.cycles.detail.delete')}
              </button>
            )}
          </div>
        </div>
        <div
          className="flex items-center gap-2 text-flow-meta"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          <span>{formatShortDate(cycle.startDate, i18n.language)}</span>
          <ArrowRight size={10} className="rtl:rotate-180" />
          <span>{formatShortDate(cycle.endDate, i18n.language)}</span>
        </div>
        {cycle.goal && (
          <p className="text-flow-body" dir="auto" style={{ color: 'var(--flow-text)' }}>
            {cycle.goal}
          </p>
        )}
      </div>

      <div className="p-6 max-w-4xl space-y-8">
        <section aria-labelledby="cycle-progress">
          <h2 id="cycle-progress" className="sr-only">
            {t('flow.cycles.detail.progress')}
          </h2>
          <CycleProgress cycle={cycle} />
        </section>

        <section>
          <h2
            className="text-flow-meta font-semibold uppercase tracking-wider mb-3"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            {t('flow.cycles.detail.burndown')}
          </h2>
          <BurndownChart points={burndown} unitLabel={t('flow.cycles.detail.unitTasks')} />
        </section>

        {carriedOver.length > 0 && (
          <section>
            <h2
              className="text-flow-meta font-semibold uppercase tracking-wider mb-2"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flow.cycles.detail.carriedOver')}
            </h2>
            <ul className="space-y-1 text-flow-meta" style={{ color: 'var(--flow-text)' }}>
              {carriedOver.map((c) => (
                <li key={c.fromCycleId}>
                  {t('flow.cycles.detail.carriedFrom', {
                    count: c.taskIds?.length ?? 0,
                    name: c.fromCycleName ?? '',
                  })}
                </li>
              ))}
            </ul>
          </section>
        )}

        <TaskSection title={t('flow.cycles.detail.openTasks', { count: open.length })}>
          {open.map((task) => (
            <TaskRow key={task.id} task={task} />
          ))}
        </TaskSection>
        {cycle.status !== 'completed' && open.length > 0 && (
          <p className="text-flow-meta -mt-6" style={{ color: 'var(--flow-text-dim)' }}>
            {nextCycle
              ? t('flow.cycles.detail.carryHint', { name: nextCycle.name })
              : t('flow.cycles.detail.carryHintBacklog')}
          </p>
        )}
        <TaskSection title={t('flow.cycles.detail.doneTasks', { count: finished.length })}>
          {finished.map((task) => (
            <TaskRow key={task.id} task={task} />
          ))}
        </TaskSection>
      </div>

      <EditCycleDialog
        cycle={editing ? cycle : null}
        pending={updateCycle.isPending}
        onOpenChange={setEditing}
        onSave={(edit) =>
          updateCycle.mutate(
            { id: cycle.id, data: edit },
            {
              onSuccess: () => {
                toast.success(t('flowToasts.cycle.updated'));
                setEditing(false);
              },
            }
          )
        }
      />
      <ConfirmDialog
        open={confirmingDelete}
        onOpenChange={setConfirmingDelete}
        title={t('flow.cycles.detail.deleteTitle')}
        description={t('flow.cycles.detail.deleteDescription', {
          name: cycle.name,
          count: members.length,
        })}
        confirmLabel={t('flowCommon.delete')}
        cancelLabel={t('flowCommon.cancel')}
        pending={deleteCycle.isPending}
        onConfirm={() =>
          deleteCycle.mutate(cycle.id, {
            onSuccess: () => {
              toast.success(t('flowToasts.cycle.deleted'));
              setConfirmingDelete(false);
              void navigate({ to: '/flow/cycles' });
            },
          })
        }
      />
    </div>
  );
};

const TaskSection: React.FC<{ title: string; children: React.ReactNode[] }> = ({
  title,
  children,
}) => (
  <section>
    <h2
      className="text-flow-meta font-semibold uppercase tracking-wider mb-2"
      style={{ color: 'var(--flow-text-muted)' }}
    >
      {title}
    </h2>
    {children.length > 0 ? (
      <div
        className="rounded-lg overflow-hidden"
        style={{ border: '1px solid var(--flow-border)' }}
      >
        {children}
      </div>
    ) : null}
  </section>
);

export default CycleDetail;
