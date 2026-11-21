import React, { useRef } from 'react';
import { useTranslation } from 'react-i18next';
import { ArrowDown, ArrowUp } from 'lucide-react';
import { cn } from '../../../shared/lib/utils';
import {
  clampColumnWidth,
  columnWidth,
  listColumn,
  MAX_COLUMN_WIDTH,
  type ListColumnId,
  type ListColumnPrefs,
} from './listColumns';
import { GRID_CELL, GRID_ROW } from './gridKeyboard';

/** How far one arrow key press moves a column edge. */
const RESIZE_STEP = 16;
/** Below this the header has no room for its name; it is still read out. */
const LABEL_MIN_WIDTH = 48;

const gridCell = { [GRID_CELL]: '' };

function ResizeHandle({
  column,
  label,
  width,
  onResize,
}: {
  column: ListColumnId;
  label: string;
  width: number;
  onResize: (column: ListColumnId, width: number) => void;
}) {
  const { t } = useTranslation();
  const drag = useRef<{ x: number; width: number; sign: number } | null>(null);
  const min = listColumn(column).minWidth;

  return (
    <div
      role="separator"
      aria-orientation="vertical"
      aria-label={t('flow.projectList.resizeColumn', { column: label })}
      aria-valuenow={width}
      aria-valuemin={min}
      aria-valuemax={MAX_COLUMN_WIDTH}
      tabIndex={0}
      className="flow-focus absolute inset-y-0 -end-2 z-10 w-3 cursor-col-resize touch-none"
      onClick={(event) => event.stopPropagation()}
      onDoubleClick={(event) => {
        event.stopPropagation();
        onResize(column, listColumn(column).width);
      }}
      onPointerDown={(event) => {
        event.preventDefault();
        event.stopPropagation();
        event.currentTarget.setPointerCapture?.(event.pointerId);
        // Dragging toward the inline end widens, whichever way the page reads.
        const rtl = getComputedStyle(event.currentTarget).direction === 'rtl';
        drag.current = { x: event.clientX, width, sign: rtl ? -1 : 1 };
      }}
      onPointerMove={(event) => {
        const start = drag.current;
        if (!start) return;
        onResize(column, start.width + (event.clientX - start.x) * start.sign);
      }}
      onPointerUp={() => {
        drag.current = null;
      }}
      onPointerCancel={() => {
        drag.current = null;
      }}
      onKeyDown={(event) => {
        const rtl = getComputedStyle(event.currentTarget).direction === 'rtl';
        const grow = rtl ? 'ArrowLeft' : 'ArrowRight';
        const shrink = rtl ? 'ArrowRight' : 'ArrowLeft';
        if (event.key !== grow && event.key !== shrink) return;
        // The grid would otherwise take the arrow as a move to the next cell.
        event.preventDefault();
        event.stopPropagation();
        onResize(column, width + (event.key === grow ? RESIZE_STEP : -RESIZE_STEP));
      }}
    />
  );
}

function HeaderCell({
  column,
  prefs,
  onSort,
  onResize,
}: {
  column: ListColumnId;
  prefs: ListColumnPrefs;
  onSort: (column: ListColumnId) => void;
  onResize: (column: ListColumnId, width: number) => void;
}) {
  const { t } = useTranslation();
  const def = listColumn(column);
  const label = t(def.labelKey);
  const width = columnWidth(prefs, column);
  const sorted = prefs.sort?.column === column ? prefs.sort.direction : null;
  const fill = column === 'title';
  const SortIcon = sorted === 'desc' ? ArrowDown : ArrowUp;

  return (
    <div
      role="columnheader"
      {...gridCell}
      data-column={column}
      tabIndex={-1}
      aria-sort={sorted === 'asc' ? 'ascending' : sorted === 'desc' ? 'descending' : undefined}
      className={cn(
        'flow-focus-inset relative flex min-w-0 items-center rounded-sm',
        fill ? 'flex-1' : 'shrink-0'
      )}
      style={fill ? { minWidth: width } : { width }}
      onKeyDown={(event) => {
        if (event.target !== event.currentTarget || event.key !== 'Enter') return;
        event.preventDefault();
        onSort(column);
      }}
    >
      <button
        type="button"
        tabIndex={-1}
        title={label}
        onClick={() => onSort(column)}
        className={cn(
          'flex min-w-0 flex-1 items-center gap-1 uppercase tracking-wider',
          def.alignEnd && 'justify-end'
        )}
      >
        <span className={cn('truncate', width < LABEL_MIN_WIDTH && !fill && 'sr-only')}>
          {label}
        </span>
        {sorted && <SortIcon size={11} aria-hidden="true" className="shrink-0" />}
      </button>
      {!fill && (
        <ResizeHandle
          column={column}
          label={label}
          width={width}
          onResize={(id, next) => onResize(id, clampColumnWidth(id, next))}
        />
      )}
    </div>
  );
}

/**
 * The grid's header row. It sits inside the scrolling area and sticks to its
 * top, so it scrolls sideways with the cells it names.
 */
export function ListHeader({
  prefs,
  onSort,
  onResize,
}: {
  prefs: ListColumnPrefs;
  onSort: (column: ListColumnId) => void;
  onResize: (column: ListColumnId, width: number) => void;
}) {
  const { t } = useTranslation();
  return (
    <div
      role="row"
      {...{ [GRID_ROW]: '' }}
      className="sticky top-0 z-20 flex items-center gap-3 px-4 py-2 text-flow-meta font-semibold"
      style={{
        color: 'var(--flow-text-dim)',
        borderBottom: '1px solid var(--flow-border)',
        background: 'var(--flow-surface)',
      }}
    >
      <div
        role="columnheader"
        {...gridCell}
        tabIndex={-1}
        aria-label={t('flow.projectList.toggleDoneLabel')}
        className="flow-focus-inset h-4 w-4 shrink-0 rounded-sm"
      />
      {prefs.columns.map((column) => (
        <HeaderCell
          key={column}
          column={column}
          prefs={prefs}
          onSort={onSort}
          onResize={onResize}
        />
      ))}
      <div
        role="columnheader"
        {...gridCell}
        tabIndex={-1}
        aria-label={t('flow.projectList.moreOptions')}
        className="flow-focus-inset h-4 w-5 shrink-0 rounded-sm"
      />
    </div>
  );
}
