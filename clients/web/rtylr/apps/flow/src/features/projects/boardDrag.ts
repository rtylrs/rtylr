/**
 * Geometry and threshold helpers for the board's pointer-based card drag.
 *
 * HTML5 drag-and-drop never fired on touch, so the board was mouse-only (B1).
 * Pointer Events cover mouse, touch and pen through one code path.
 *
 * These are separated from the React wiring so the parts that decide *whether*
 * a gesture is a drag, and *where* it landed, can be tested without a real
 * pointer.
 */

/** Movement, in CSS px, before a press counts as a drag rather than a click. */
export const DRAG_THRESHOLD_PX = 6;

export interface ColumnRect {
  id: string;
  left: number;
  right: number;
  top: number;
  bottom: number;
}

/**
 * Whether the pointer has travelled far enough to be a drag.
 *
 * Uses squared distance so the comparison needs no square root, and treats the
 * threshold as radial rather than per-axis: a diagonal nudge of 5px on each axis
 * is 7px of travel and should count, which two independent axis checks would
 * miss.
 */
export function exceedsDragThreshold(dx: number, dy: number): boolean {
  return dx * dx + dy * dy >= DRAG_THRESHOLD_PX * DRAG_THRESHOLD_PX;
}

/**
 * The column under a viewport point, or null when the pointer is outside all of
 * them.
 *
 * Bounds are inclusive on both edges so a pointer resting exactly on the seam
 * between two columns resolves to the first rather than to nothing. Rects are
 * measured at drag start: the board scrolls, so re-measuring per move would be
 * correct but costs a layout read on every pointermove.
 */
export function columnIdAtPoint(rects: readonly ColumnRect[], x: number, y: number): string | null {
  for (const r of rects) {
    if (x >= r.left && x <= r.right && y >= r.top && y <= r.bottom) return r.id;
  }
  return null;
}

/**
 * Where a task should land, or null when the move is a no-op.
 *
 * Dropping a card back on its own column is not a move: forwarding it would
 * spend a request and an optimistic update to change nothing.
 */
export function resolveDropTarget(
  targetColumnId: string | null,
  sourceColumnId: string | null
): string | null {
  if (targetColumnId === null) return null;
  if (targetColumnId === sourceColumnId) return null;
  return targetColumnId;
}
