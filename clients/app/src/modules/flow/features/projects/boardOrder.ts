/**
 * Ordering helpers for the board's live card drag.
 *
 * Kept apart from the React wiring, like `boardDrag.ts`, so the decisions that
 * matter (which slot a pointer is over, what order a drop stores, how saved and
 * unsaved cards interleave) can be tested without a pointer or a DOM.
 */

/** A card's vertical centre, measured in viewport px when the drag began. */
export interface CardSlot {
  id: string;
  mid: number;
}

/** Where the dragged card would land: a column and a position among its cards. */
export interface DropSlot {
  columnId: string;
  index: number;
}

/**
 * Cards in the order the board shows them.
 *
 * `sortOrder` is 1-based once a column has been arranged by hand and 0 until
 * then. A card nobody has placed yet keeps the order the server sent (it sorts
 * by sort_order, then creation time) and sits above the placed ones, which is
 * also where a card created after an arrangement lands. Array#sort is stable,
 * so ties keep input order.
 */
export function orderBoardTasks<T extends { sortOrder?: number }>(tasks: readonly T[]): T[] {
  return [...tasks].sort((a, b) => (a.sortOrder ?? 0) - (b.sortOrder ?? 0));
}

/**
 * The position the dragged card should take among the other cards of a column:
 * how many of them sit above the pointer.
 *
 * Slots are measured once, at drag start, and the dragged card is skipped.
 * Using the start-of-drag geometry rather than live rects keeps the answer
 * independent of the neighbours' layout animation, so a card never flickers
 * between two positions while it is mid-slide.
 */
export function insertionIndex(
  slots: readonly CardSlot[],
  draggedId: string,
  pointerY: number
): number {
  let index = 0;
  for (const slot of slots) {
    if (slot.id === draggedId) continue;
    if (slot.mid < pointerY) index += 1;
  }
  return index;
}

/**
 * A column's card ids after the dragged card is dropped at `index` among the
 * others. This is the list sent to the server, so it holds every card of the
 * column, dragged one included.
 */
export function orderedIdsAfterDrop(
  columnIds: readonly string[],
  draggedId: string,
  index: number
): string[] {
  const others = columnIds.filter((id) => id !== draggedId);
  const at = Math.max(0, Math.min(index, others.length));
  return [...others.slice(0, at), draggedId, ...others.slice(at)];
}

/** Whether dropping at `slot` leaves the card exactly where it started. */
export function isSamePlace(
  source: { columnId: string; index: number },
  slot: DropSlot | null
): boolean {
  return slot === null || (slot.columnId === source.columnId && slot.index === source.index);
}

/**
 * Applies a stored order to cached tasks, for the optimistic update after a
 * drop: listed cards get `sortOrder` = position (1-based) and the moved card is
 * placed in its destination column. Everything else is returned untouched, and
 * the same array is returned when nothing changed so cache subscribers do not
 * re-render for nothing.
 */
export function applyBoardMove<T extends { id: string | number; sortOrder?: unknown }>(
  tasks: readonly T[],
  move: { taskId: string; columnId: string; orderedTaskIds?: readonly string[] }
): T[] {
  const position = new Map<string, number>();
  move.orderedTaskIds?.forEach((id, i) => position.set(String(id), i + 1));
  let changed = false;
  const next = tasks.map((task) => {
    const id = String(task.id);
    const isMoved = id === move.taskId;
    const order = position.get(id);
    if (!isMoved && order === undefined) return task;
    changed = true;
    const patch: Record<string, unknown> = {};
    if (isMoved) patch.sectionId = move.columnId;
    if (order !== undefined) patch.sortOrder = order;
    return { ...task, ...patch };
  });
  return changed ? next : (tasks as T[]);
}
