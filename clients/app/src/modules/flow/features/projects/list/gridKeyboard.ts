import type React from 'react';

/**
 * Arrow key movement for the List's grid, worked out from the DOM rather than
 * from data: rows are grouped, collapsible, windowed and nest their subtasks,
 * and the DOM is the one place that already has them in on-screen order.
 *
 * A row is any element with `data-grid-row`; a cell is any element inside one
 * with `data-grid-cell`. Cells are focused directly (tabindex -1), so a click
 * also sets where the arrows start from.
 */
export const GRID_ROW = 'data-grid-row';
export const GRID_CELL = 'data-grid-cell';

const isTyping = (target: HTMLElement): boolean =>
  target.isContentEditable || /^(INPUT|TEXTAREA|SELECT)$/.test(target.tagName);

function cellsOf(row: Element): HTMLElement[] {
  return Array.from(row.querySelectorAll<HTMLElement>(`[${GRID_CELL}]`)).filter(
    // A subtask row sits inside its parent's wrapper, not inside its row.
    (cell) => cell.closest(`[${GRID_ROW}]`) === row
  );
}

function focusCell(cell: HTMLElement | undefined): boolean {
  if (!cell) return false;
  cell.focus();
  cell.scrollIntoView?.({ block: 'nearest', inline: 'nearest' });
  return true;
}

/**
 * Returns the grid's keydown handler. `column` remembers the column the user
 * last moved to, so passing through a group header (one wide cell) and on to
 * the next task lands back in that column.
 */
export function handleGridKeyDown(
  event: React.KeyboardEvent<HTMLElement>,
  column: { current: number }
): void {
  const grid = event.currentTarget;
  const target = event.target as HTMLElement;
  // Keys from a menu or picker opened by a cell reach here through the React
  // tree but are not in the grid's DOM: those belong to the menu.
  if (event.defaultPrevented || !grid.contains(target) || isTyping(target)) return;
  if (event.altKey || event.metaKey) return;

  const cell = target.closest<HTMLElement>(`[${GRID_CELL}]`);
  const row = cell?.closest(`[${GRID_ROW}]`);
  if (!cell || !row) return;

  const cells = cellsOf(row);
  const index = cells.indexOf(cell);
  const rtl = getComputedStyle(grid).direction === 'rtl';
  const forward = rtl ? 'ArrowLeft' : 'ArrowRight';
  const back = rtl ? 'ArrowRight' : 'ArrowLeft';

  let moved = false;
  switch (event.key) {
    case forward:
    case back: {
      const next = index + (event.key === forward ? 1 : -1);
      if (next >= 0 && next < cells.length) {
        column.current = next;
        moved = focusCell(cells[next]);
      }
      break;
    }
    case 'Home':
    case 'End': {
      if (event.ctrlKey) {
        const rows = Array.from(grid.querySelectorAll(`[${GRID_ROW}]`));
        const edge = event.key === 'Home' ? rows[0] : rows[rows.length - 1];
        const edgeCells = edge ? cellsOf(edge) : [];
        moved = focusCell(edgeCells[Math.min(column.current, edgeCells.length - 1)]);
      } else {
        const next = event.key === 'Home' ? 0 : cells.length - 1;
        column.current = next;
        moved = focusCell(cells[next]);
      }
      break;
    }
    case 'ArrowDown':
    case 'ArrowUp': {
      if (event.ctrlKey || event.shiftKey) return;
      const rows = Array.from(grid.querySelectorAll(`[${GRID_ROW}]`));
      // One wide cell (a group header) says nothing about the column.
      if (cells.length > 1) column.current = index;
      let at = rows.indexOf(row);
      const step = event.key === 'ArrowDown' ? 1 : -1;
      for (at += step; at >= 0 && at < rows.length; at += step) {
        const nextCells = cellsOf(rows[at]);
        if (nextCells.length === 0) continue;
        moved = focusCell(nextCells[Math.min(column.current, nextCells.length - 1)]);
        break;
      }
      break;
    }
    default:
      return;
  }
  if (moved) event.preventDefault();
}
