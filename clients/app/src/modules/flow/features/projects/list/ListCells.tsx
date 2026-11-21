import React, { useEffect, useMemo, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Calendar } from 'lucide-react';
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from '@voxire/ui';
import { useFlowStore, type Task } from '../../../shared/stores/flowStore';
import type { ApiStatus } from '../../../shared/services/flowApi';
import { cn, formatDueDate, formatShortDate, getInitials } from '../../../shared/lib/utils';
import PriorityIcon from '../../../shared/components/PriorityIcon';
import { FLOW_MENU_SURFACE, FlowMenuItem } from '../../../shared/components/FlowSelect';
import { LabelChip } from '../../../shared/components/LabelChip';
import { AssigneeAvatar } from '../../../shared/components/AssigneeAvatar';
import { useLabels } from '../../tasks/useLabels';
import { useProjectStatuses } from '../useStatuses';
import { displayStatusName, sortStatuses } from '../statusUtils';
import { PRIORITY_LABEL_KEY, PRIORITY_ORDER } from '../taskGroups';
import { listColumn, resolveTaskStatus, type ListColumnId } from './listColumns';
import { GRID_CELL } from './gridKeyboard';
import { useListCellEdit } from './useListCellEdit';

interface CellProps {
  task: Task;
  width: number;
  /** False for viewers: every cell then reads only. */
  canEdit: boolean;
}

type Close = (refocus: boolean) => void;

interface FrameProps {
  column: ListColumnId;
  width: number;
  canEdit: boolean;
  children: React.ReactNode;
  /** Rendered while editing: a menu opens over the value, an input replaces it. */
  editor?: (close: Close) => React.ReactNode;
  replaceWhileEditing?: boolean;
}

/**
 * One cell of the grid. Focusable for arrow key movement; Enter (or F2, or a
 * click) starts editing when the column is editable and the user may edit.
 * A read-only cell lets the click through to the row, which opens the task.
 */
function CellFrame({
  column,
  width,
  canEdit,
  children,
  editor,
  replaceWhileEditing = false,
}: FrameProps) {
  const def = listColumn(column);
  const ref = useRef<HTMLDivElement>(null);
  const [editing, setEditing] = useState(false);
  const editable = canEdit && def.editable && !!editor;
  const close: Close = (refocus) => {
    setEditing(false);
    if (!refocus) return;
    ref.current?.focus();
    // A menu's focus trap holds focus until it unmounts, so try once more
    // after it has gone, unless the user has since moved focus elsewhere.
    setTimeout(() => {
      const active = document.activeElement;
      if (!active || active === document.body) ref.current?.focus();
    }, 0);
  };

  return (
    <div
      ref={ref}
      role="gridcell"
      {...{ [GRID_CELL]: '' }}
      data-column={column}
      tabIndex={-1}
      aria-readonly={def.editable && !editable ? true : undefined}
      className={cn(
        'flow-focus-inset relative flex h-5 min-w-0 shrink-0 items-center rounded-sm text-flow-meta',
        def.alignEnd && 'justify-end',
        editable && 'cursor-pointer hover:bg-[var(--flow-hover)]'
      )}
      style={{ width, color: 'var(--flow-text-dim)' }}
      onClick={(event) => {
        // Clicks from a menu's portal arrive through the React tree; they are
        // the menu's, not the cell's.
        if (!editable || !event.currentTarget.contains(event.target as Node)) return;
        event.stopPropagation();
        if (!editing) setEditing(true);
      }}
      onKeyDown={(event) => {
        if (!editable || editing || event.target !== event.currentTarget) return;
        if (event.key === 'Enter' || event.key === 'F2') {
          event.preventDefault();
          event.stopPropagation();
          setEditing(true);
        }
      }}
    >
      {editing && replaceWhileEditing ? null : children}
      {editing && editor?.(close)}
    </div>
  );
}

/**
 * A menu anchored over the cell. Radix would hand focus back to its trigger,
 * an anchor that unmounts with the menu, so the cell takes it back instead
 * (see CellFrame's close), Escape included.
 */
function CellMenu({
  label,
  onClose,
  children,
}: {
  label: string;
  onClose: () => void;
  children: React.ReactNode;
}) {
  return (
    <DropdownMenu open onOpenChange={(open) => !open && onClose()}>
      <DropdownMenuTrigger asChild>
        <span aria-label={label} tabIndex={-1} className="pointer-events-none absolute inset-0" />
      </DropdownMenuTrigger>
      <DropdownMenuContent
        align="start"
        aria-label={label}
        onCloseAutoFocus={(event) => event.preventDefault()}
        className="max-h-[260px] min-w-[180px] overflow-y-auto"
        style={FLOW_MENU_SURFACE}
      >
        {children}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

/**
 * A text-like editor. Enter saves, Escape drops the change, and leaving the
 * field saves too, as a spreadsheet does.
 */
function CellInput({
  type,
  initial,
  label,
  close,
  onCommit,
  ...rest
}: {
  type: 'date' | 'number';
  initial: string;
  label: string;
  close: Close;
  onCommit: (value: string) => void;
} & Pick<React.InputHTMLAttributes<HTMLInputElement>, 'min' | 'step' | 'inputMode'>) {
  const ref = useRef<HTMLInputElement>(null);
  const done = useRef(false);
  useEffect(() => {
    ref.current?.focus();
  }, []);

  const finish = (save: boolean, refocus: boolean): void => {
    if (done.current) return;
    done.current = true;
    const value = ref.current?.value ?? initial;
    if (save && value !== initial) onCommit(value);
    close(refocus);
  };

  return (
    <input
      ref={ref}
      type={type}
      defaultValue={initial}
      aria-label={label}
      className="h-5 w-full min-w-0 rounded-sm px-1 text-flow-meta"
      style={{
        background: 'var(--flow-input-bg)',
        border: '1px solid var(--flow-border-control)',
        color: 'var(--flow-text)',
      }}
      onKeyDown={(event) => {
        if (event.key !== 'Enter' && event.key !== 'Escape') return;
        event.preventDefault();
        event.stopPropagation();
        finish(event.key === 'Enter', true);
      }}
      onBlur={() => finish(true, false)}
      {...rest}
    />
  );
}

const dateOnly = (value: string | undefined): string => (value ? value.slice(0, 10) : '');

function PriorityCell({ task, width, canEdit }: CellProps) {
  const { t } = useTranslation();
  return (
    <CellFrame
      column="priority"
      width={width}
      canEdit={canEdit}
      editor={(close) => <PriorityEditor task={task} onClose={() => close(true)} />}
    >
      <PriorityIcon priority={task.priority} size={12} />
      {width >= 96 && (
        <span className="ms-1.5 truncate" style={{ color: 'var(--flow-text)' }}>
          {t(PRIORITY_LABEL_KEY[task.priority])}
        </span>
      )}
    </CellFrame>
  );
}

function PriorityEditor({ task, onClose }: { task: Task; onClose: () => void }) {
  const { t } = useTranslation();
  const edit = useListCellEdit();
  return (
    <CellMenu label={t(listColumn('priority').labelKey)} onClose={onClose}>
      {PRIORITY_ORDER.map((priority) => (
        <FlowMenuItem
          key={priority}
          label={t(PRIORITY_LABEL_KEY[priority])}
          icon={<PriorityIcon priority={priority} size={12} />}
          selected={priority === task.priority}
          onSelect={() => {
            if (priority !== task.priority) edit.update(task.id, { priority });
          }}
        />
      ))}
    </CellMenu>
  );
}

const StatusDot = ({ color }: { color: string }) => (
  <span
    aria-hidden="true"
    className="h-2 w-2 shrink-0 rounded-full"
    style={{ background: color }}
  />
);

function StatusCell({ task, width, canEdit }: CellProps) {
  const { t } = useTranslation();
  const { data: statuses } = useProjectStatuses(task.projectId);
  const ordered = useMemo(() => sortStatuses(statuses ?? []), [statuses]);
  const current = resolveTaskStatus(task, ordered);
  const name = current ? displayStatusName(current.name, t) : (task.statusName ?? '');
  return (
    <CellFrame
      column="status"
      width={width}
      canEdit={canEdit && ordered.length > 0}
      editor={(close) => (
        <StatusEditor
          task={task}
          statuses={ordered}
          currentId={current?.id}
          onClose={() => close(true)}
        />
      )}
    >
      <StatusDot color={current?.color || task.statusColor || 'var(--flow-text-dim)'} />
      <span className="ms-1.5 truncate" style={{ color: 'var(--flow-text)' }}>
        {name}
      </span>
    </CellFrame>
  );
}

function StatusEditor({
  task,
  statuses,
  currentId,
  onClose,
}: {
  task: Task;
  statuses: readonly ApiStatus[];
  currentId: string | undefined;
  onClose: () => void;
}) {
  const { t } = useTranslation();
  const edit = useListCellEdit();
  return (
    <CellMenu label={t(listColumn('status').labelKey)} onClose={onClose}>
      {statuses.map((status) => (
        <FlowMenuItem
          key={status.id}
          label={displayStatusName(status.name, t)}
          icon={<StatusDot color={status.color} />}
          selected={status.id === currentId}
          onSelect={() => {
            if (status.id !== currentId) edit.setStatus(task.id, status);
          }}
        />
      ))}
    </CellMenu>
  );
}

function AssigneeCell({ task, width, canEdit }: CellProps) {
  const { t } = useTranslation();
  const members = useFlowStore((state) => state.members);
  // Subtasks can arrive without the display name; the id still resolves.
  const name =
    task.assignee ||
    (task.assigneeId ? members.find((m) => String(m.id) === task.assigneeId)?.name : '') ||
    '';
  const initials = task.assigneeInitials || (name ? getInitials(name) : '');
  return (
    <CellFrame
      column="assignee"
      width={width}
      canEdit={canEdit}
      editor={(close) => <AssigneeEditor task={task} name={name} onClose={() => close(true)} />}
    >
      {name ? (
        // The shared avatar: the person's chosen picture or colour, and their
        // name on hover (#1650).
        <AssigneeAvatar name={name} initials={initials} seed={task.assigneeId} size={20} />
      ) : (
        <span
          role="img"
          aria-label={t('flow.quickCreate.unassigned')}
          className="h-5 w-5 rounded-full shrink-0"
          style={{ background: 'var(--accent-muted)' }}
        />
      )}
      {width >= 96 && name && (
        <span className="ms-1.5 truncate" style={{ color: 'var(--flow-text)' }}>
          {name}
        </span>
      )}
    </CellFrame>
  );
}

function AssigneeEditor({
  task,
  name,
  onClose,
}: {
  task: Task;
  name: string;
  onClose: () => void;
}) {
  const { t } = useTranslation();
  const edit = useListCellEdit();
  const members = useFlowStore((state) => state.members);
  const current = task.assigneeId || '';
  const options = [
    { id: '', name: t('flow.quickCreate.unassigned'), initials: '' },
    // Someone assigned who has since left the member list stays choosable.
    ...(current && !members.some((m) => String(m.id) === current)
      ? [{ id: current, name, initials: task.assigneeInitials }]
      : []),
    ...members.map((m) => ({ id: String(m.id), name: m.name, initials: m.initials })),
  ];
  return (
    <CellMenu label={t(listColumn('assignee').labelKey)} onClose={onClose}>
      {options.map((option) => (
        <FlowMenuItem
          key={option.id || '_none'}
          label={option.name}
          selected={option.id === current}
          onSelect={() => {
            if (option.id === current) return;
            const assigned = option.id ? option.name : '';
            edit.update(
              task.id,
              { assigneeId: option.id },
              {
                assigneeName: assigned,
                assignee: assigned,
                assigneeInitials: option.id ? option.initials || getInitials(assigned) : '',
              }
            );
          }}
        />
      ))}
    </CellMenu>
  );
}

function DateCell({ task, width, canEdit, field }: CellProps & { field: 'startDate' | 'dueDate' }) {
  const { t } = useTranslation();
  const value = (field === 'dueDate' ? task.dueDate : task.startDate) ?? '';
  return (
    <CellFrame
      column={field}
      width={width}
      canEdit={canEdit}
      replaceWhileEditing
      editor={(close) => <DateEditor task={task} field={field} value={value} close={close} />}
    >
      <span className="flex items-center gap-1">
        {field === 'dueDate' && <Calendar size={11} aria-hidden="true" />}
        {field === 'dueDate'
          ? formatDueDate(value, t)
          : formatShortDate(value, t('flowCommon.dateLocale'))}
      </span>
    </CellFrame>
  );
}

function DateEditor({
  task,
  field,
  value,
  close,
}: {
  task: Task;
  field: 'startDate' | 'dueDate';
  value: string;
  close: Close;
}) {
  const { t } = useTranslation();
  const edit = useListCellEdit();
  return (
    <CellInput
      type="date"
      initial={dateOnly(value)}
      label={t(listColumn(field).labelKey)}
      close={close}
      // Empty clears the date.
      onCommit={(next) => edit.update(task.id, { [field]: next })}
    />
  );
}

function EstimateCell({ task, width, canEdit }: CellProps) {
  const { t } = useTranslation();
  const hours = task.estimate != null && task.estimate > 0 ? task.estimate : null;
  return (
    <CellFrame
      column="estimate"
      width={width}
      canEdit={canEdit}
      replaceWhileEditing
      editor={(close) => <EstimateEditor task={task} hours={hours} close={close} />}
    >
      <span className="tabular-nums">
        {hours == null ? '-' : t('flow.projectList.estimateValue', { hours })}
      </span>
    </CellFrame>
  );
}

function EstimateEditor({
  task,
  hours,
  close,
}: {
  task: Task;
  hours: number | null;
  close: Close;
}) {
  const { t } = useTranslation();
  const edit = useListCellEdit();
  return (
    <CellInput
      type="number"
      min={0}
      step="any"
      inputMode="decimal"
      initial={hours == null ? '' : String(hours)}
      label={t(listColumn('estimate').labelKey)}
      close={close}
      onCommit={(next) => {
        // Empty clears the estimate, as 0 does in the task panel.
        const value = next.trim() === '' ? 0 : Number(next);
        if (!Number.isFinite(value) || value < 0) return;
        edit.update(task.id, { estimatedHours: value });
      }}
    />
  );
}

function LabelsCell({ task, width, canEdit }: CellProps) {
  const { data: labels } = useLabels();
  const ids = task.labelIds ?? [];
  const shown = (labels ?? []).filter((label) => ids.includes(String(label.id)));
  return (
    <CellFrame
      column="labels"
      width={width}
      canEdit={canEdit}
      editor={(close) => <LabelsEditor task={task} onClose={() => close(true)} />}
    >
      <span className="flex min-w-0 items-center gap-1 overflow-hidden">
        {shown.map((label) => (
          <LabelChip
            key={label.id}
            small
            label={{ id: String(label.id), name: label.name, color: label.color }}
          />
        ))}
      </span>
    </CellFrame>
  );
}

function LabelsEditor({ task, onClose }: { task: Task; onClose: () => void }) {
  const { t } = useTranslation();
  const edit = useListCellEdit();
  const { data: labels = [] } = useLabels();
  // Held here so several toggles in one opening build on each other rather
  // than each on the value the row had when the menu opened.
  const [selected, setSelected] = useState<string[]>(() => (task.labelIds ?? []).map(String));
  return (
    <CellMenu label={t(listColumn('labels').labelKey)} onClose={onClose}>
      {labels.length === 0 && (
        <p className="px-3 py-2 text-flow-meta" style={{ color: 'var(--flow-text-muted)' }}>
          {t('flow.projectList.noLabels')}
        </p>
      )}
      {labels.map((label) => {
        const id = String(label.id);
        const on = selected.includes(id);
        return (
          <FlowMenuItem
            key={id}
            checkable
            label={label.name}
            icon={<StatusDot color={label.color} />}
            selected={on}
            onSelect={() => {
              const next = on ? selected.filter((value) => value !== id) : [...selected, id];
              setSelected(next);
              edit.setLabels(task.id, next);
            }}
          />
        );
      })}
    </CellMenu>
  );
}

function PlainCell({
  column,
  width,
  children,
}: {
  column: ListColumnId;
  width: number;
  children: React.ReactNode;
}) {
  return (
    <CellFrame column={column} width={width} canEdit={false}>
      <span className="truncate">{children}</span>
    </CellFrame>
  );
}

/** Every column but the title, which the row draws itself. */
export function ListCell({ column, ...props }: CellProps & { column: ListColumnId }) {
  const { t } = useTranslation();
  switch (column) {
    case 'priority':
      return <PriorityCell {...props} />;
    case 'status':
      return <StatusCell {...props} />;
    case 'assignee':
      return <AssigneeCell {...props} />;
    case 'startDate':
    case 'dueDate':
      return <DateCell {...props} field={column} />;
    case 'estimate':
      return <EstimateCell {...props} />;
    case 'labels':
      return <LabelsCell {...props} />;
    case 'key':
      return (
        <PlainCell column="key" width={props.width}>
          {props.task.identifier}
        </PlainCell>
      );
    case 'created':
      return (
        <PlainCell column="created" width={props.width}>
          {formatShortDate(props.task.createdAt, t('flowCommon.dateLocale'))}
        </PlainCell>
      );
    case 'title':
      return null;
  }
}
