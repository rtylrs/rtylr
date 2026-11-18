import React, { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { ChevronDown, ChevronRight, UsersRound } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { Avatar } from '@voxire/ui';
import { useTasks } from './useTasks';
import { useFlowStore, type Task } from '../../shared/stores/flowStore';
import { apiTaskToTask } from '../../shared/lib/taskMapper';
import { TaskRow } from '../../shared/components/TaskRow';
import { VirtualList } from '../../../../shared/components/VirtualList';
import { hasQueryFailed } from '../../shared/lib/queryFailureUtils';
import { ListLoadFailed } from '../../shared/components/ListLoadFailed';
import { useCurrentUserId } from '../../shared/lib/currentUser';
import { initialsFromName } from '../../shared/lib/memberUtils';
import FilterBar from '../../shared/filters/FilterBar';
import ShowCompletedToggle from '../../shared/filters/ShowCompletedToggle';
import {
  includesClosed,
  isClosedTask,
  useShowCompleted,
} from '../../shared/filters/useShowCompleted';
import { assigneeField, orderPeople, type FilterPerson } from '../../shared/filters/filterFields';
import { UNASSIGNED } from '../../shared/filters/taskFilters';
import { PRIORITY_ORDER } from '../projects/taskGroups';
import { useCrossProjectTaskFilters } from './useMyTaskFilters';

interface PersonSection {
  id: string;
  name: string;
  tasks: Task[];
}

/** Earliest due first, undated last, then by priority. */
function byUrgency(a: Task, b: Task): number {
  const da = a.dueDate ? a.dueDate.split('T')[0] : '';
  const db = b.dueDate ? b.dueDate.split('T')[0] : '';
  if (da !== db) {
    if (da === '') return 1;
    if (db === '') return -1;
    return da < db ? -1 : 1;
  }
  return PRIORITY_ORDER.indexOf(a.priority) - PRIORITY_ORDER.indexOf(b.priority);
}

/** Everyone who holds a task, by id, with the name the task carries. */
export function peopleOnTasks(tasks: readonly Task[]): FilterPerson[] {
  const byId = new Map<string, string>();
  for (const task of tasks) {
    if (task.assigneeId && !byId.has(task.assigneeId)) {
      byId.set(task.assigneeId, task.assignee || task.assigneeId);
    }
  }
  return [...byId].map(([id, name]) => ({ id, name }));
}

/** One section per person, the current user first, Unassigned last. */
export function sectionsByPerson(
  tasks: readonly Task[],
  people: readonly FilterPerson[],
  currentUserId: string | null,
  unassignedLabel: string
): PersonSection[] {
  const sections = orderPeople(people, currentUserId).map((p) => ({
    id: p.id,
    name: p.name,
    tasks: tasks.filter((task) => task.assigneeId === p.id).sort(byUrgency),
  }));
  sections.push({
    id: UNASSIGNED,
    name: unassignedLabel,
    tasks: tasks.filter((task) => !task.assigneeId).sort(byUrgency),
  });
  return sections.filter((section) => section.tasks.length > 0);
}

/**
 * Team tasks: every task you can see across the business, one section per
 * person, with the same filter bar as a project board. The list call without a
 * project returns only what the server's visibility policy lets the caller
 * see, so this page shows nobody anything they could not already open.
 */
const TeamTasks: React.FC = () => {
  const { t } = useTranslation();
  const currentUserId = useCurrentUserId();
  const selectedTaskIds = useFlowStore((state) => state.selectedTaskIds);
  const tasksQuery = useTasks({});
  const loadFailed = hasQueryFailed(tasksQuery);
  const isLoading = tasksQuery.isPending && !loadFailed;
  const tasks = useMemo(
    () => (tasksQuery.data ?? []).map(apiTaskToTask).filter((task) => !task.archived),
    [tasksQuery.data]
  );

  const people = useMemo(() => peopleOnTasks(tasks), [tasks]);
  const personField = useMemo(
    () => [assigneeField(people, currentUserId, t)],
    [people, currentUserId, t]
  );
  const { state, fields, matchingIds } = useCrossProjectTaskFilters(
    'team-tasks',
    tasks,
    personField
  );
  const [showCompleted, setShowCompleted] = useShowCompleted('team-tasks');
  const includeClosed = includesClosed(showCompleted, state.filters.statuses);

  const inScope = useMemo(
    () => tasks.filter((task) => includeClosed || !isClosedTask(task)),
    [tasks, includeClosed]
  );
  const shown = useMemo(
    () => (matchingIds === null ? inScope : inScope.filter((task) => matchingIds.has(task.id))),
    [inScope, matchingIds]
  );
  const sections = useMemo(
    () => sectionsByPerson(shown, people, currentUserId, t('flow.filters.unassigned')),
    [shown, people, currentUserId, t]
  );

  const [collapsed, setCollapsed] = useState<ReadonlySet<string>>(new Set());
  const toggle = (id: string): void =>
    setCollapsed((current) => {
      const next = new Set(current);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });

  return (
    <div className="flex h-full flex-col">
      <HelmetContainer title={`${t('flow.shell.teamTasks')} · Flow`} />
      <div
        className="portal-flow-page-header flex items-center justify-between gap-4 px-6 py-5 shrink-0"
        style={{ borderBottom: '1px solid var(--flow-border)' }}
      >
        <div>
          <h1 className="text-flow-title font-semibold" style={{ color: 'var(--flow-text)' }}>
            {t('flow.shell.teamTasks')}
          </h1>
          {!loadFailed && (
            <p className="text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
              {isLoading
                ? t('flow.myIssues.loading')
                : t('flow.teamTasks.summary', { tasks: inScope.length, people: people.length })}
            </p>
          )}
        </div>
      </div>

      {tasks.length > 0 && !loadFailed && (
        <FilterBar
          state={state}
          currentUserId={currentUserId}
          fields={fields}
          people={people}
          visible={shown.length}
          total={inScope.length}
          allowDim={false}
          extra={<ShowCompletedToggle pressed={showCompleted} onChange={setShowCompleted} />}
        />
      )}

      <div className="relative flex-1 overflow-y-auto">
        <div className="space-y-5 p-4">
          {tasks.length > 0 && sections.length === 0 && (
            <p
              className="px-1 py-6 text-center text-flow-meta"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flow.filters.noMatches')}
            </p>
          )}
          {sections.map((section) => {
            const isCollapsed = collapsed.has(section.id);
            const listId = `team-tasks-${section.id}`;
            return (
              <section key={section.id} data-team-section={section.id}>
                <h2>
                  <button
                    type="button"
                    onClick={() => toggle(section.id)}
                    aria-expanded={!isCollapsed}
                    aria-controls={listId}
                    className="flow-focus flex min-h-11 w-full items-center gap-2 px-1 text-start text-flow-meta font-semibold"
                    style={{ color: 'var(--flow-text)' }}
                  >
                    {isCollapsed ? <ChevronRight size={14} /> : <ChevronDown size={14} />}
                    {section.id === UNASSIGNED ? (
                      <span
                        className="flex h-6 w-6 items-center justify-center rounded-full"
                        style={{ border: '1px dashed var(--flow-border)' }}
                        aria-hidden="true"
                      />
                    ) : (
                      <Avatar
                        size="xs"
                        initials={initialsFromName(section.name)}
                        seed={section.id}
                        aria-hidden
                      />
                    )}
                    {section.id === currentUserId
                      ? t('flow.filters.you', { name: section.name })
                      : section.name}
                    <span
                      className="rounded-sm px-1.5 py-0.5 text-flow-meta font-medium"
                      style={{ background: 'var(--flow-hover)', color: 'var(--flow-text-muted)' }}
                    >
                      {section.tasks.length}
                    </span>
                  </button>
                </h2>
                <VirtualList
                  id={listId}
                  hidden={isCollapsed}
                  className="overflow-hidden rounded-xl"
                  style={{ background: 'var(--flow-card)', border: '1px solid var(--flow-border)' }}
                  count={isCollapsed ? 0 : section.tasks.length}
                  estimateSize={41}
                  renderRow={(index, rowProps) => (
                    <TaskRow
                      key={section.tasks[index].id}
                      rowProps={rowProps}
                      task={section.tasks[index]}
                      showProject
                      selected={selectedTaskIds.includes(section.tasks[index].id)}
                    />
                  )}
                />
              </section>
            );
          })}
        </div>

        {loadFailed && !isLoading && (
          <div className="absolute inset-0 z-10 flex items-center justify-center p-4">
            <div className="w-full max-w-md">
              <ListLoadFailed onRetry={() => void tasksQuery.refetch()} />
            </div>
          </div>
        )}

        {tasks.length === 0 && !isLoading && !loadFailed && (
          <div className="pointer-events-none absolute inset-0 z-10 flex items-center justify-center p-4">
            <div
              className="portal-flow-empty pointer-events-auto w-full max-w-md px-6 py-7 text-center"
              style={{ background: 'var(--flow-surface)' }}
            >
              <div className="portal-flow-empty-icon">
                <UsersRound size={24} />
              </div>
              <h2 className="text-flow-title font-semibold" style={{ color: 'var(--flow-text)' }}>
                {t('flow.teamTasks.emptyTitle')}
              </h2>
              <p className="mt-2 text-flow-meta" style={{ color: 'var(--flow-text-muted)' }}>
                {t('flow.teamTasks.emptyDescription')}
              </p>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default TeamTasks;
