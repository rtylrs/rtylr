import React, { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Flag, TreePalm } from 'lucide-react';
import type { MonthCell } from '../../shared/lib/monthGrid';
import {
  chunkWeeks,
  hiddenPerDay,
  laneCount,
  layoutWeek,
  taskRange,
  type DayRange,
  type WeekSegment,
} from '../../shared/lib/calendarSpans';
import { addDays } from '../../shared/lib/dateUtils';
import { withAlpha, onTint } from '../../shared/lib/colorUtils';
import type { CalendarView } from './useCalendarView';
import type { CalendarScheduling } from './useCalendarScheduling';
import type { LeaveDay, PlanningCycle, PlanningMilestone } from './useCalendarPlanning';

/** What the grid needs from a task. Both calendars pass ApiTask. */
export interface GridTask {
  id: string | number;
  title: string;
  dueDate?: string | null;
  startDate?: string | null;
}

/** Height of the row holding the day number, above the bars. */
const HEADER_PX = 32;
const TASK_LANE_PX = 22;
const BAND_LANE_PX = 18;
/** Bars shown per week before the rest fold into "+N more". */
const TASK_LANES: Record<CalendarView, number> = { month: 3, week: 12 };
/** Cycle and milestone rows shown per week. */
const BAND_LANES: Record<CalendarView, number> = { month: 2, week: 4 };

type BandItem =
  | { kind: 'cycle'; key: string; cycle: PlanningCycle }
  | { kind: 'milestone'; key: string; milestone: PlanningMilestone };

const MILESTONE_TONE: Record<PlanningMilestone['status'], string> = {
  planning: 'var(--flow-text-muted)',
  active: 'var(--accent)',
  at_risk: 'var(--flow-warn-fg)',
  completed: 'var(--flow-status-approved)',
};

export interface PlanningGridProps<T extends GridTask> {
  cells: readonly MonthCell[];
  view: CalendarView;
  tasks: readonly T[];
  cycles: readonly PlanningCycle[];
  milestones: readonly PlanningMilestone[];
  leave: ReadonlyMap<string, LeaveDay>;
  today: string;
  /** A task's colour, or undefined for the plain card surface. */
  tone: (task: T) => string | undefined;
  /** A cycle's colour; the info tone when absent. */
  cycleTone?: (cycle: PlanningCycle) => string | undefined;
  leading?: (task: T) => React.ReactNode;
  /** Whether the signed-in user may change this task's dates. */
  canEdit: (task: T) => boolean;
  onOpen: (task: T) => void;
  /** Present on a calendar that creates a task from a day. */
  onCreate?: (cell: MonthCell) => void;
  scheduling: Pick<CalendarScheduling<T>, 'active' | 'begin' | 'end' | 'drop'>;
}

function isRtl(): boolean {
  return typeof document !== 'undefined' && document.documentElement.dir === 'rtl';
}

/**
 * The day of a bar the pointer picked it up on, so a moved bar keeps the grabbed
 * day under the pointer. Falls back to the segment's first day when the bar
 * has no width to measure (a test DOM).
 */
function grabbedDay(event: React.DragEvent<HTMLElement>, firstDay: string, span: number): string {
  const rect = event.currentTarget.getBoundingClientRect();
  if (rect.width <= 0 || span <= 1) return firstDay;
  const x = isRtl() ? rect.right - event.clientX : event.clientX - rect.left;
  const offset = Math.min(span - 1, Math.max(0, Math.floor(x / (rect.width / span))));
  return addDays(firstDay, offset);
}

function startDrag(event: React.DragEvent<HTMLElement>, taskId: string): void {
  // Firefox will not start a drag without data; nothing reads it back.
  event.dataTransfer?.setData('text/plain', taskId);
  if (event.dataTransfer) event.dataTransfer.effectAllowed = 'move';
  event.stopPropagation();
}

function gridRows(band: number, tasks: number): string {
  const rows = [
    ...Array.from({ length: band }, () => `${BAND_LANE_PX}px`),
    ...Array.from({ length: tasks }, () => `${TASK_LANE_PX}px`),
  ];
  return rows.length > 0 ? rows.join(' ') : 'none';
}

/**
 * The calendar body shared by Plan > Calendar and the project calendar.
 *
 * Each week is a row of seven day cells with the bars laid over it: cycle and
 * milestone bars in a band at the top, then the tasks, each task running from
 * its start date to its due date. The cells take the drops and the create
 * clicks; while a drag is in flight the bars stop taking the pointer so the
 * cell underneath sees it. The grid follows the document direction, so in
 * Arabic the week runs right to left with no extra work here.
 */
export function PlanningGrid<T extends GridTask>({
  cells,
  view,
  tasks,
  cycles,
  milestones,
  leave,
  today,
  tone,
  cycleTone,
  leading,
  canEdit,
  onOpen,
  onCreate,
  scheduling,
}: PlanningGridProps<T>): React.ReactElement {
  const { t, i18n } = useTranslation();
  const [hoverDay, setHoverDay] = useState<string | null>(null);
  const dragging = scheduling.active !== null;

  const weekdays = useMemo(() => {
    const formatter = new Intl.DateTimeFormat(i18n.language, { weekday: 'short' });
    return cells.slice(0, 7).map((cell) => formatter.format(cell.date));
  }, [cells, i18n.language]);

  const taskItems = useMemo(() => {
    const items: { item: T; range: DayRange }[] = [];
    for (const task of tasks) {
      const range = taskRange(task);
      if (range) items.push({ item: task, range });
    }
    return items;
  }, [tasks]);

  const bandItems = useMemo(() => {
    const items: { item: BandItem; range: DayRange }[] = [];
    for (const cycle of cycles) {
      items.push({ item: { kind: 'cycle', key: `c-${cycle.id}`, cycle }, range: cycle.range });
    }
    for (const milestone of milestones) {
      items.push({
        item: { kind: 'milestone', key: `m-${milestone.id}`, milestone },
        range: { start: milestone.day, end: milestone.day },
      });
    }
    return items;
  }, [cycles, milestones]);

  const weeks = useMemo(() => chunkWeeks(cells), [cells]);
  const longDate = (date: Date): string =>
    date.toLocaleDateString(i18n.language, { month: 'long', day: 'numeric', year: 'numeric' });

  const renderTask = (segment: WeekSegment<T>, week: readonly MonthCell[], row: number) => {
    const task = segment.item;
    const id = String(task.id);
    const editable = canEdit(task);
    const color = tone(task);
    const firstDay = week[segment.col].iso;
    return (
      <div
        key={`t-${id}`}
        className={['relative min-w-0 px-0.5', dragging ? '' : 'pointer-events-auto'].join(' ')}
        style={{ gridColumn: `${segment.col + 1} / span ${segment.span}`, gridRow: row }}
      >
        <button
          type="button"
          title={task.title}
          dir="auto"
          draggable={editable}
          onDragStart={
            editable
              ? (event) => {
                  startDrag(event, id);
                  scheduling.begin({
                    taskId: id,
                    kind: 'move',
                    grabDay: grabbedDay(event, firstDay, segment.span),
                  });
                }
              : undefined
          }
          onDragEnd={editable ? scheduling.end : undefined}
          onClick={(event) => {
            event.stopPropagation();
            onOpen(task);
          }}
          className={[
            'flex h-5 w-full items-center gap-1 px-1.5 text-start text-flow-meta truncate hover:brightness-95',
            segment.startsHere ? 'rounded-s-sm' : '',
            segment.endsHere ? 'rounded-e-sm' : '',
            editable ? 'cursor-grab active:cursor-grabbing' : 'cursor-pointer',
          ].join(' ')}
          style={{
            // The tone appears once and both the tint and the foreground read it
            // back, which also makes it observable in tests: happy-dom drops any
            // color-mix() set through the style property (rtylr#1249).
            ...(color ? { ['--tone' as string]: color } : {}),
            background: color ? withAlpha('var(--tone)', 13) : 'var(--flow-card)',
            color: color ? onTint('var(--tone)') : 'var(--flow-text)',
          }}
        >
          {leading?.(task)}
          {/* dir="auto" so a right-to-left title truncates at its own end
              rather than clipping the start of it. */}
          <span dir="auto" className="truncate">
            {task.title}
          </span>
        </button>
        {editable && segment.startsHere && (
          <span
            aria-hidden="true"
            title={t('flow.calendar.resizeStart')}
            draggable
            data-edge="start"
            onDragStart={(event) => {
              startDrag(event, id);
              scheduling.begin({ taskId: id, kind: 'start' });
            }}
            onDragEnd={scheduling.end}
            className="absolute inset-y-0 start-0 w-1.5 cursor-ew-resize rounded-s-sm hover:bg-[var(--flow-border-focus)]"
          />
        )}
        {editable && segment.endsHere && (
          <span
            aria-hidden="true"
            title={t('flow.calendar.resizeEnd')}
            draggable
            data-edge="end"
            onDragStart={(event) => {
              startDrag(event, id);
              scheduling.begin({ taskId: id, kind: 'end' });
            }}
            onDragEnd={scheduling.end}
            className="absolute inset-y-0 end-0 w-1.5 cursor-ew-resize rounded-e-sm hover:bg-[var(--flow-border-focus)]"
          />
        )}
      </div>
    );
  };

  const renderBand = (segment: WeekSegment<BandItem>, row: number) => {
    const entry = segment.item;
    const placement = { gridColumn: `${segment.col + 1} / span ${segment.span}`, gridRow: row };
    const rounding = [
      segment.startsHere ? 'rounded-s-sm' : '',
      segment.endsHere ? 'rounded-e-sm' : '',
    ].join(' ');
    if (entry.kind === 'cycle') {
      const color = cycleTone?.(entry.cycle) ?? 'var(--flow-info-fg)';
      const label = t('flow.calendar.cycleTitle', { name: entry.cycle.name });
      return (
        <div
          key={entry.key}
          className={['min-w-0 px-0.5', dragging ? '' : 'pointer-events-auto'].join(' ')}
          style={placement}
        >
          <div
            title={label}
            aria-label={label}
            data-cycle={entry.cycle.id}
            className={`flex h-4 items-center px-1.5 text-flow-meta truncate ${rounding}`}
            style={{
              ['--tone' as string]: color,
              background: withAlpha('var(--tone)', 22),
              color: onTint('var(--tone)'),
            }}
          >
            {/* The name once, where the cycle starts or where the week picks it up. */}
            <span dir="auto" className="truncate">
              {entry.cycle.name}
            </span>
          </div>
        </div>
      );
    }
    const color = MILESTONE_TONE[entry.milestone.status] ?? 'var(--accent)';
    const label = t('flow.calendar.milestoneTitle', { name: entry.milestone.name });
    return (
      <div
        key={entry.key}
        className={['min-w-0 px-0.5', dragging ? '' : 'pointer-events-auto'].join(' ')}
        style={placement}
      >
        <div
          title={label}
          aria-label={label}
          data-milestone={entry.milestone.id}
          className="flex h-4 items-center gap-1 px-1 text-flow-meta truncate rounded-sm"
          style={{ ['--tone' as string]: color, color: onTint('var(--tone)') }}
        >
          <Flag size={10} className="shrink-0" style={{ color: 'var(--tone)' }} />
          <span dir="auto" className="truncate">
            {entry.milestone.name}
          </span>
        </div>
      </div>
    );
  };

  return (
    <div className="flex flex-1 flex-col min-h-0 min-w-0">
      <div className="grid grid-cols-7 gap-px shrink-0">
        {weekdays.map((day) => (
          <div
            key={day}
            className="py-1.5 text-center text-flow-meta font-semibold uppercase"
            style={{ color: 'var(--flow-text-dim)' }}
          >
            {day}
          </div>
        ))}
      </div>
      <div
        className="flex flex-1 flex-col gap-px overflow-auto rounded-md"
        style={{ background: 'var(--flow-border)' }}
      >
        {weeks.map((week) => {
          const weekStart = week[0].iso;
          const taskSegments = layoutWeek(weekStart, taskItems);
          const bandSegments = layoutWeek(weekStart, bandItems);
          const maxTasks = TASK_LANES[view];
          const maxBand = BAND_LANES[view];
          const bandLanes = Math.min(laneCount(bandSegments), maxBand);
          const taskLanes = Math.min(laneCount(taskSegments), maxTasks);
          const hidden = hiddenPerDay(taskSegments, maxTasks);
          const moreRow = hidden.some((count) => count > 0) ? 1 : 0;
          const bodyPx = bandLanes * BAND_LANE_PX + (taskLanes + moreRow) * TASK_LANE_PX;
          const minHeight = Math.max(view === 'week' ? 320 : 100, HEADER_PX + bodyPx + 6);
          return (
            <div
              key={weekStart}
              className="relative grid flex-1 grid-cols-7 gap-px"
              style={{ minHeight }}
            >
              {week.map((cell) => {
                const isToday = cell.iso === today;
                const leaveDay = leave.get(cell.iso);
                const isDropTarget = dragging && hoverDay === cell.iso;
                const creatable = onCreate !== undefined && cell.inMonth;
                const surface = cell.inMonth ? 'var(--flow-bg)' : 'var(--flow-surface)';
                return (
                  <div
                    key={cell.iso}
                    data-day={cell.iso}
                    data-leave={leaveDay ? leaveDay.people : undefined}
                    onClick={creatable ? () => onCreate(cell) : undefined}
                    onDragOver={(event) => {
                      if (!dragging) return;
                      event.preventDefault();
                      if (event.dataTransfer) event.dataTransfer.dropEffect = 'move';
                      if (hoverDay !== cell.iso) setHoverDay(cell.iso);
                    }}
                    onDragLeave={() => setHoverDay((day) => (day === cell.iso ? null : day))}
                    onDrop={(event) => {
                      event.preventDefault();
                      setHoverDay(null);
                      scheduling.drop(cell.iso);
                    }}
                    className={[
                      'group flex min-w-0 flex-col p-1.5',
                      creatable ? 'cursor-pointer' : '',
                    ].join(' ')}
                    style={{
                      background: isDropTarget ? withAlpha('var(--accent)', 12) : surface,
                      // Leave is a hatch rather than a fill, so it reads under the
                      // bars and next to today's highlight without hiding either.
                      ...(leaveDay
                        ? {
                            backgroundImage: `repeating-linear-gradient(135deg, ${withAlpha('var(--flow-warn-fg)', 10)} 0 6px, transparent 6px 12px)`,
                          }
                        : {}),
                    }}
                  >
                    <div
                      className="flex items-center justify-between gap-1"
                      style={{ height: HEADER_PX - 6 }}
                    >
                      <span
                        className="flex h-5 min-w-5 items-center justify-center rounded-full px-1 text-flow-meta font-semibold"
                        style={{
                          background: isToday ? 'var(--flow-active)' : 'transparent',
                          color: isToday
                            ? 'var(--accent-on-muted)'
                            : cell.inMonth
                              ? 'var(--flow-text-muted)'
                              : 'var(--flow-text-dim)',
                        }}
                      >
                        {cell.date.getDate()}
                      </span>
                      {leaveDay && (
                        // An icon and a number, because a month cell has no room
                        // for the sentence; the sentence is the tooltip and what
                        // a screen reader hears.
                        <span
                          className="flex min-w-0 items-center gap-0.5 text-flow-meta"
                          title={t('flow.calendar.leaveCount', { count: leaveDay.people })}
                          style={{ color: 'var(--flow-warn-fg)' }}
                        >
                          <TreePalm size={11} aria-hidden="true" className="shrink-0" />
                          <span aria-hidden="true">{leaveDay.people}</span>
                          <span className="sr-only">
                            {t('flow.calendar.leaveCount', { count: leaveDay.people })}
                          </span>
                        </span>
                      )}
                      {creatable && (
                        <button
                          type="button"
                          onClick={(event) => {
                            event.stopPropagation();
                            onCreate(cell);
                          }}
                          aria-label={t('flow.calendar.addTaskLabel', {
                            date: longDate(cell.date),
                          })}
                          className="flex h-[44px] w-[44px] shrink-0 items-center justify-center rounded-sm text-flow-meta opacity-0 group-hover:opacity-60 group-focus-within:opacity-60 focus:opacity-100 transition-opacity focus:focus-visible:ring-1 focus-visible:ring-[var(--accent)] md:h-6 md:w-auto md:px-1"
                          style={{ color: 'var(--flow-text-dim)' }}
                        >
                          {t('flow.calendar.addHint')}
                        </button>
                      )}
                    </div>
                  </div>
                );
              })}

              <div
                className="pointer-events-none absolute inset-x-0 grid grid-cols-7 gap-x-px gap-y-0.5"
                style={{
                  top: HEADER_PX,
                  gridTemplateRows: gridRows(bandLanes, taskLanes + moreRow),
                }}
              >
                {bandSegments
                  .filter((segment) => segment.lane < maxBand)
                  .map((segment) => renderBand(segment, segment.lane + 1))}
                {taskSegments
                  .filter((segment) => segment.lane < maxTasks)
                  .map((segment) => renderTask(segment, week, bandLanes + segment.lane + 1))}
                {hidden.map((count, col) =>
                  count > 0 ? (
                    <span
                      key={`more-${col}`}
                      onClick={(event) => event.stopPropagation()}
                      className={[
                        'px-1.5 text-flow-meta truncate',
                        dragging ? '' : 'pointer-events-auto',
                      ].join(' ')}
                      style={{
                        gridColumn: col + 1,
                        gridRow: bandLanes + taskLanes + 1,
                        color: 'var(--flow-text-dim)',
                      }}
                    >
                      {t('flow.calendar.moreCount', { count })}
                    </span>
                  ) : null
                )}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
