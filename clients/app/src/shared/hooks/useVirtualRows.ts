import { useLayoutEffect, useRef, useState } from 'react';
import { useVirtualizer } from '@tanstack/react-virtual';
import { scrollParent } from '../lib/scrollParent';

/** Lists shorter than this render every row; windowing them buys nothing. */
export const VIRTUAL_ROWS_FROM = 80;

interface VirtualRowsOptions {
  /** Expected row height in px. Rows are measured once they render. */
  estimateSize: number;
  /** Rows rendered beyond each edge of the visible area. */
  overscan?: number;
  /** Space between rows, for a flex or grid list with a gap. */
  gap?: number;
  /** The row count from which rows are windowed. */
  from?: number;
  /**
   * Header rows above the body, for a table: rows then carry aria-rowindex and
   * the table aria-rowcount, so assistive tech knows where it is in the whole.
   */
  headerRows?: number;
}

export interface RowProps {
  'data-index'?: number;
  'aria-rowindex'?: number;
  ref?: (node: Element | null) => void;
}

export interface VirtualRows {
  /** Attach to the element that holds the rows: a tbody, or the list itself. */
  listRef: (node: HTMLElement | null) => void;
  /** The rows to render, in order. Every index when not windowed. */
  indexes: number[];
  /** Space standing in for the unrendered rows above and below. */
  before: number;
  after: number;
  /** Spread on each rendered row so its real height is measured. */
  rowProps: (index: number) => RowProps;
  windowed: boolean;
}

/**
 * Renders only the rows near the visible part of a long list or table.
 *
 * Lists here scroll with their page (the portal's main column) rather than in
 * a box of their own, so this watches the nearest scrolling ancestor and
 * offsets by where the list starts inside it. `before` and `after` keep the
 * page its full height, so whatever follows the list (a load-more sentinel, a
 * footer) stays put. Rows are measured as they render, so mixed heights are
 * fine.
 *
 * In a table, columns hold their width while windowed, so they don't jump as
 * rows come and go. The lock lets go when the rows on screen need more room
 * than it allows, and when the scroller changes width.
 * Where nothing can be measured (no scrolling ancestor, as in jsdom) every row
 * renders.
 */
export function useVirtualRows(
  count: number,
  { estimateSize, overscan = 8, gap = 0, from = VIRTUAL_ROWS_FROM, headerRows }: VirtualRowsOptions
): VirtualRows {
  const [list, setList] = useState<HTMLElement | null>(null);
  // undefined until looked for; null when there is none and every row renders.
  const [scroller, setScroller] = useState<HTMLElement | null | undefined>(undefined);
  const [margin, setMargin] = useState(0);
  // Each column's locked width, in header-cell order.
  const columns = useRef<number[]>([]);
  const long = count >= from;

  useLayoutEffect(() => {
    if (!list || !long) {
      setScroller(undefined);
      return;
    }
    const el = scrollParent(list);
    setScroller(el && el.clientHeight > 0 ? el : null);
  }, [list, long]);

  useLayoutEffect(() => {
    if (!list || !scroller) return;
    let width = scroller.clientWidth;
    const measure = () => {
      // Where the first row starts: inside the list's own border and padding.
      const { borderTopWidth, paddingTop } = getComputedStyle(list);
      const top =
        list.getBoundingClientRect().top +
        (parseFloat(borderTopWidth) || 0) +
        (parseFloat(paddingTop) || 0) -
        scroller.getBoundingClientRect().top -
        scroller.clientTop +
        scroller.scrollTop;
      setMargin((current) => (Math.abs(current - top) < 1 ? current : top));
      if (scroller.clientWidth !== width) {
        width = scroller.clientWidth;
        columns.current = [];
        for (const cell of headerCells(list)) cell.style.minWidth = '';
      }
    };
    measure();
    if (typeof ResizeObserver === 'undefined') return;
    // The page above the list can grow (stats load, chips appear) without the
    // list itself changing, so watch the scroller's content too.
    const observer = new ResizeObserver(measure);
    observer.observe(scroller);
    for (const child of scroller.children) observer.observe(child);
    return () => observer.disconnect();
  }, [list, scroller]);

  const windowed = long && !!scroller;
  const virtualizer = useVirtualizer({
    count,
    enabled: windowed,
    getScrollElement: () => scroller ?? null,
    estimateSize: () => estimateSize,
    overscan,
    gap,
    scrollMargin: margin,
  });
  const items = windowed ? virtualizer.getVirtualItems() : [];

  useLayoutEffect(() => {
    if (!list) return;
    const table = list.parentElement;
    const cells = headerCells(list);
    if (!windowed) {
      columns.current = [];
      for (const cell of cells) cell.style.minWidth = '';
      if (headerRows !== undefined) table?.removeAttribute('aria-rowcount');
      return;
    }
    if (headerRows !== undefined) table?.setAttribute('aria-rowcount', String(count + headerRows));
    if (!(table instanceof HTMLTableElement) || cells.length === 0 || items.length === 0) return;

    // Lock each column at its width as shown. Shown widths fill the table's
    // box exactly, so there is no spare room left to reshuffle between
    // columns as rows come and go. When the rows now on screen need more than
    // the locked widths allow, the table overflows: unlock, and lock again at
    // what these rows need.
    if (
      columns.current.length > 0 &&
      table.scrollWidth > (table.parentElement?.clientWidth ?? 0) + 1
    ) {
      columns.current = [];
      for (const cell of cells) cell.style.minWidth = '';
    }
    columns.current = cells.map((cell, i) =>
      Math.max(cell.getBoundingClientRect().width, columns.current[i] ?? 0)
    );
    columns.current.forEach((width, i) => (cells[i].style.minWidth = `${width}px`));
  });

  if (!windowed) {
    return {
      listRef: setList,
      // Still looking for the scroller: hold the space, render nothing yet.
      indexes: long && scroller === undefined ? [] : Array.from({ length: count }, (_, i) => i),
      before: 0,
      after: long && scroller === undefined ? count * (estimateSize + gap) : 0,
      rowProps: () => ({}),
      windowed: false,
    };
  }

  // In a list with a gap, each spacer sits next to one gap of its own.
  const first = items[0];
  const last = items[items.length - 1];
  return {
    listRef: setList,
    indexes: items.map((item) => item.index),
    before: first && first.index > 0 ? first.start - margin - gap : 0,
    after:
      last && last.index < count - 1 ? virtualizer.getTotalSize() - (last.end - margin) - gap : 0,
    rowProps: (index) => ({
      'data-index': index,
      ref: virtualizer.measureElement,
      ...(headerRows !== undefined ? { 'aria-rowindex': index + headerRows + 1 } : {}),
    }),
    windowed: true,
  };
}

/** A table body's header cells, or none for a plain list. */
function headerCells(list: HTMLElement): HTMLTableCellElement[] {
  if (!(list instanceof HTMLTableSectionElement)) return [];
  const table = list.parentElement;
  if (!(table instanceof HTMLTableElement)) return [];
  return Array.from(
    table.querySelectorAll<HTMLTableCellElement>(':scope > thead > tr:first-child > th')
  );
}
