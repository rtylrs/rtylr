/**
 * Placement decisions for the sidebar's live row drag, kept apart from the
 * React wiring so they can be tested without a drag event.
 */

/** Where a dragged row would land: a group (or the top level) and the row it goes before. */
export interface ItemPlacement {
  /** Undefined means the top-level list. */
  groupId?: string;
  /** Undefined means the end of the list. */
  beforeItemId?: string;
}

/**
 * The placement for a row dragged over `overItemId`: before it when the pointer
 * is in its upper half, after it otherwise.
 *
 * `listIds` is the list being hovered, as stored (the dragged row may be in it).
 * "After" is expressed as "before the next row", because that is what a move is
 * stored as; the last row has no next, so the result is the end of the list.
 */
export function placementOver(
  listIds: readonly string[],
  draggedId: string,
  overItemId: string,
  after: boolean,
  groupId?: string
): ItemPlacement {
  const others = listIds.filter((id) => id !== draggedId);
  const at = others.indexOf(overItemId);
  if (at < 0) return { groupId };
  return { groupId, beforeItemId: after ? others[at + 1] : others[at] };
}

export function samePlacement(a: ItemPlacement | undefined, b: ItemPlacement | undefined): boolean {
  if (a === b) return true;
  if (!a || !b) return false;
  return a.groupId === b.groupId && a.beforeItemId === b.beforeItemId;
}

/**
 * Whether a placement stays inside the list the row is already in. Only those
 * moves are previewed live: a row that changes list is unmounted and remounted
 * by React, and a browser drag does not survive its source node being replaced.
 */
export function isWithinSameList(
  sourceGroupId: string | undefined,
  placement: ItemPlacement
): boolean {
  return placement.groupId === sourceGroupId;
}

/** Group order after dragging `draggedId` onto `targetId`: it takes the target's place. */
export function reorderGroupIds(
  order: readonly string[],
  draggedId: string,
  targetId: string
): string[] {
  const from = order.indexOf(draggedId);
  const to = order.indexOf(targetId);
  if (from < 0 || to < 0 || from === to) return [...order];
  const next = [...order];
  next.splice(from, 1);
  next.splice(to, 0, draggedId);
  return next;
}
