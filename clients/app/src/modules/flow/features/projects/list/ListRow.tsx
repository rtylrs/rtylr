import React, { useEffect, useMemo, useState } from 'react';
import { motion } from 'motion/react';
import { Check, ChevronRight, MoreHorizontal } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import type { Task } from '../../../shared/stores/flowStore';
import type { ApiTask } from '../../../shared/services/flowApi';
import { cn } from '../../../shared/lib/utils';
import { apiTaskToTask } from '../../../shared/lib/taskMapper';
import { useOpenTask } from '../../../shared/lib/useOpenTask';
import { useCanEditTask } from '../../../shared/lib/currentUser';
import { TaskContextMenu, type ContextMenuState } from '../../../shared/components/TaskContextMenu';
import EpicChip from '../../../shared/components/EpicChip';
import type { RowProps } from '../../../../../shared/hooks/useVirtualRows';
import { useToggleDone } from '../../tasks/useToggleDone';
import { useTaskDetail } from '../../tasks/useTasks';
import { columnWidth, DEFAULT_LIST_PREFS } from './listColumns';
import { useListGrid } from './listGridContext';
import { GRID_CELL, GRID_ROW } from './gridKeyboard';
import { ListCell } from './ListCells';

/** Subtasks of subtasks open too, down to this depth. */
const MAX_DEPTH = 4;
const INDENT = 20;

interface TaskRowProps {
  task: Task;
  depth?: number;
  /** From a windowed list (see VirtualList). */
  rowProps?: RowProps;
  /** Faded because it does not match the filters ("Dim instead of hide"). */
  dimmed?: boolean;
}

const gridCell = { [GRID_CELL]: '' };

/** A cell holding one control: Enter or Space on the cell presses it. */
function ControlCell({ className, children }: { className: string; children: React.ReactNode }) {
  return (
    <div
      role="gridcell"
      {...gridCell}
      tabIndex={-1}
      className={cn('flow-focus-inset flex shrink-0 items-center rounded-sm', className)}
      onKeyDown={(event) => {
        if (event.target !== event.currentTarget) return;
        if (event.key !== 'Enter' && event.key !== ' ') return;
        event.preventDefault();
        event.currentTarget.querySelector<HTMLButtonElement>('button:not(:disabled)')?.click();
      }}
    >
      {children}
    </div>
  );
}

function TitleCell({
  task,
  depth,
  width,
  checked,
  expanded,
  onToggleExpanded,
}: {
  task: Task;
  depth: number;
  width: number;
  checked: boolean;
  expanded: boolean;
  onToggleExpanded: () => void;
}) {
  const { t } = useTranslation();
  const openTask = useOpenTask();
  const open = () => openTask(task.id);
  const hasChildren = task.subtaskCount > 0 && depth < MAX_DEPTH;

  return (
    <div
      role="gridcell"
      {...gridCell}
      data-column="title"
      tabIndex={-1}
      className="flow-focus-inset flex min-w-0 flex-1 items-center gap-1.5 rounded-sm"
      style={{ minWidth: width, paddingInlineStart: depth * INDENT }}
      onKeyDown={(event) => {
        if (event.target !== event.currentTarget || event.key !== 'Enter') return;
        event.preventDefault();
        open();
      }}
    >
      {hasChildren && (
        <button
          type="button"
          aria-expanded={expanded}
          aria-label={t(
            expanded ? 'flow.projectList.hideSubtasks' : 'flow.projectList.showSubtasks',
            { count: task.subtaskCount }
          )}
          onClick={(event) => {
            event.stopPropagation();
            onToggleExpanded();
          }}
          className="flex shrink-0 items-center gap-0.5 rounded-sm px-0.5 text-flow-meta tabular-nums hover:bg-[var(--flow-hover)]"
          style={{ color: 'var(--flow-text-dim)' }}
        >
          <ChevronRight
            size={12}
            aria-hidden="true"
            className={cn('transition-transform', expanded ? 'rotate-90' : 'rtl:rotate-180')}
          />
          {task.subtaskCount}
        </button>
      )}
      <button
        type="button"
        // A title in another script truncates at its own end.
        dir="auto"
        aria-label={t('flowCommon.taskRow.openTask', {
          identifier: task.identifier,
          title: task.title,
        })}
        onClick={(event) => {
          event.stopPropagation();
          open();
        }}
        onKeyDown={(event) => {
          if (event.key !== 'Enter' && event.key !== ' ') return;
          event.preventDefault();
          event.stopPropagation();
          open();
        }}
        className={cn(
          'min-w-0 flex-1 truncate rounded-sm text-start text-flow-body font-medium',
          checked ? 'line-through' : ''
        )}
        style={{ color: checked ? 'var(--flow-text-dim)' : 'var(--flow-text)' }}
      >
        {task.title}
      </button>
      <EpicChip task={task} />
    </div>
  );
}

/**
 * One task as a grid row. The Done box, the title and the More button are
 * always there; the cells between them are the columns the user picked.
 * Clicking the row opens the task, except on a cell the user can edit.
 */
export const TaskRow: React.FC<TaskRowProps> = ({ task, depth = 0, rowProps, dimmed = false }) => {
  // Windowed rows mount as they scroll into view, so they skip the entrance
  // and layout animations a short list uses.
  const windowed = rowProps?.['data-index'] !== undefined;
  const { t } = useTranslation();
  const grid = useListGrid();
  const [checked, setChecked] = useState(task.status === 'done');
  const [ownExpanded, setOwnExpanded] = useState(false);
  const [contextMenu, setContextMenu] = useState<ContextMenuState | null>(null);
  const toggleDone = useToggleDone(task);
  const openTask = useOpenTask();
  const canEdit = useCanEditTask(task.projectId);
  const prefs = { ...DEFAULT_LIST_PREFS, columns: [...grid.columns], widths: grid.widths };
  const hasChildren = task.subtaskCount > 0 && depth < MAX_DEPTH;
  const expanded = hasChildren && (grid.expanded ? grid.expanded.has(task.id) : ownExpanded);

  useEffect(() => {
    setChecked(task.status === 'done');
  }, [task.status]);

  const handleToggle = () => {
    const newChecked = !checked;
    setChecked(newChecked);
    const revert = () => setChecked(!newChecked);
    if (newChecked) toggleDone.complete(revert);
    else toggleDone.reopen(revert);
  };

  const handleMoreClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    const vw = window.innerWidth;
    const vh = window.innerHeight;
    const menuW = 160;
    const menuH = 120;
    const x = e.clientX + menuW > vw ? e.clientX - menuW : e.clientX;
    const y = e.clientY + menuH > vh ? e.clientY - menuH : e.clientY;
    setContextMenu({ x, y, taskId: task.id, projectId: task.projectId, taskTitle: task.title });
  };

  return (
    <div {...rowProps}>
      <motion.div
        role="row"
        {...{ [GRID_ROW]: '' }}
        aria-level={depth + 1}
        aria-expanded={hasChildren ? expanded : undefined}
        onClick={(event) => {
          // Clicks inside a portal (a cell's menu, the More menu) bubble here
          // through the React tree; only the row's own area opens the task.
          if (!event.currentTarget.contains(event.target as Node)) return;
          openTask(task.id);
        }}
        layout={!windowed}
        initial={windowed ? false : { opacity: 0, x: -6 }}
        animate={{ opacity: dimmed ? 0.35 : 1, x: 0 }}
        data-filter-dimmed={dimmed ? '' : undefined}
        transition={{ duration: 0.15 }}
        className="flex items-center gap-3 px-4 py-2.5 group hover:bg-[var(--flow-hover)] transition-colors cursor-pointer"
        style={{ borderBottom: '1px solid var(--flow-border)' }}
      >
        <ControlCell className="w-4">
          <button
            type="button"
            disabled={!canEdit || !toggleDone.ready}
            onClick={(event) => {
              event.stopPropagation();
              handleToggle();
            }}
            aria-label={t('flow.projectList.toggleDoneLabel')}
            aria-pressed={checked}
            className="h-4 w-4 rounded-sm shrink-0 flex items-center justify-center transition-colors border"
            style={{
              background: checked ? 'var(--accent)' : 'transparent',
              borderColor: checked ? 'var(--accent)' : 'var(--flow-border)',
            }}
          >
            {checked && <Check size={10} strokeWidth={3} color="var(--dash-on-accent)" />}
          </button>
        </ControlCell>

        {grid.columns.map((column) =>
          column === 'title' ? (
            <TitleCell
              key="title"
              task={task}
              depth={depth}
              width={columnWidth(prefs, 'title')}
              checked={checked}
              expanded={expanded}
              onToggleExpanded={() =>
                grid.expanded ? grid.toggleExpanded(task.id) : setOwnExpanded((open) => !open)
              }
            />
          ) : (
            <ListCell
              key={column}
              column={column}
              task={task}
              width={columnWidth(prefs, column)}
              canEdit={canEdit}
            />
          )
        )}

        <ControlCell className="w-5">
          <button
            type="button"
            onClick={handleMoreClick}
            aria-label={t('flow.projectList.moreOptions')}
            className="opacity-0 group-hover:opacity-100 focus-visible:opacity-100 transition-opacity h-5 w-5 rounded-sm flex items-center justify-center hover:bg-[var(--flow-hover)]"
            style={{ color: 'var(--flow-text-dim)' }}
          >
            <MoreHorizontal size={13} />
          </button>
          {contextMenu && (
            <TaskContextMenu menu={contextMenu} onClose={() => setContextMenu(null)} />
          )}
        </ControlCell>
      </motion.div>
      {expanded && <SubtaskRows parent={task} depth={depth + 1} />}
    </div>
  );
};

/**
 * A parent's subtasks, from its detail (the project list returns top level
 * tasks only). They are not filtered: a subtask shows under the parent that
 * matched, as in Jira.
 */
function SubtaskRows({ parent, depth }: { parent: Task; depth: number }) {
  const { t } = useTranslation();
  const grid = useListGrid();
  const detail = useTaskDetail(parent.id);
  const subtasks = useMemo(() => {
    // The v2 detail returns each subtask as a full task, though ApiTaskDetail
    // still types them with the old three field shape.
    const raw = (detail.data?.subtasks ?? []) as unknown as ApiTask[];
    return grid.sortRows(
      raw
        .map((subtask) =>
          apiTaskToTask({ ...subtask, projectId: subtask.projectId || parent.projectId })
        )
        .filter((subtask) => !subtask.archived)
    );
  }, [detail.data, grid, parent.projectId]);

  if (detail.isPending || detail.isError) {
    return (
      <div
        role="row"
        aria-level={depth + 1}
        style={{ borderBottom: '1px solid var(--flow-border)' }}
      >
        <div
          role="gridcell"
          className="px-4 py-2 text-flow-meta"
          style={{ paddingInlineStart: 16 + 28 + depth * INDENT, color: 'var(--flow-text-muted)' }}
        >
          {detail.isError
            ? t('flow.projectList.subtasksFailed')
            : t('flow.projectList.subtasksLoading')}
        </div>
      </div>
    );
  }

  return (
    <>
      {subtasks.map((subtask) => (
        <TaskRow key={subtask.id} task={subtask} depth={depth} />
      ))}
    </>
  );
}
