import type { Task } from '../stores/flowStore';
import type { ApiTask } from '../services/flowApi';
import { getInitials } from './utils';

function normalizePriority(
  priority: ApiTask['priority'] | 'none' | null | undefined
): Task['priority'] {
  return priority === 'none' || !priority ? 'no_priority' : priority;
}

/**
 * Normalises an ApiTask (from the server) into the richer Task shape used
 * by local store consumers and legacy components.
 *
 * Two things to watch:
 *  - The wire sends numeric IDs (proto `int64`); the Zustand Task keys them
 *    as strings (URL params, find-by-id comparisons). Always stringify.
 *  - Display fields come from the enriched backend (projectName,
 *    assigneeName, assigneeInitials, identifier, subtaskCount). Fall back to
 *    derived values when missing so the UI never renders `undefined`.
 *
 * Do NOT shallow-spread the API task (`...t`) - that leaks numeric IDs
 * through `as any` into a store typed as strings and the bug comes back.
 */
export function apiTaskToTask(t: ApiTask): Task {
  // The hand-rolled ApiTask claims `id: string` but the wire actually
  // sends `id: number` (int64). Coerce defensively until Phase 3 swaps
  // ApiTask for the proto-generated Task.
  const id = String(t.id);
  const projectId = String(t.projectId);
  const workspaceId = String(t.workspaceId);

  // Proto added `assigneeName` alongside the legacy `assignee` display name;
  // prefer whichever the backend sent.
  const assignee = (t as { assigneeName?: string }).assigneeName ?? t.assignee ?? '';

  // sectionId is the column the task lives in on the kanban board.
  // The wire sends it as a number (or null for unsectioned); the store
  // keys columns by string id, so stringify.
  const rawSection = (t as { sectionId?: number | string }).sectionId;
  const sectionId = rawSection == null || rawSection === 0 ? undefined : String(rawSection);

  return {
    id,
    projectId,
    workspaceId,
    title: t.title,
    description: t.description,
    status: t.status,
    priority: normalizePriority(t.priority),
    assignee,
    // Derived from the name that is actually rendered, not read from the wire
    // independently. `assignee` above prefers the newer `assigneeName` while
    // this field only ever carried the legacy one, so a task could show "FO"
    // beside "Mohammad Al-Homsi": two sources for one fact, disagreeing
    // (#1097, D-040). The wire value is still the fallback for an unnamed
    // assignee, where there is no name to derive anything from.
    assigneeInitials: assignee ? getInitials(assignee) : (t.assigneeInitials ?? ''),
    assigneeId: t.assigneeId != null ? String(t.assigneeId) : '',
    dueDate: t.dueDate ?? '',
    startDate: (t as { startDate?: string }).startDate,
    subtaskCount: t.subtaskCount ?? 0,
    projectName: t.projectName ?? '',
    labelIds: t.labels,
    identifier: t.identifier ?? id,
    createdAt: t.createdAt ?? new Date().toISOString(),
    updatedAt: t.updatedAt ?? new Date().toISOString(),
    estimate: t.estimate,
    // CreateTask has no effort_minutes field in the v2 contract, so the quick
    // create path sends effort as estimated_hours. Prefer the explicit field
    // when it exists and derive the legacy representation only as a fallback.
    effortMinutes:
      t.effortMinutes ??
      (t.estimatedHours != null && Number.isFinite(t.estimatedHours) && t.estimatedHours > 0
        ? Math.round(t.estimatedHours * 60)
        : undefined),
    waitingOn: t.waitingOn,
    decisionNote: t.decisionNote,
    focusPin: t.focusPin,
    sectionId,
    // `archived` is omitempty on the contract, so an unarchived task arrives
    // as `undefined`, not `false`. Normalise here so store consumers get a
    // concrete boolean and nobody is tempted to write `archived === false`
    // (which would silently match nothing).
    archived: t.archived ?? false,
    // statusId is int64-as-string on the wire already; no coercion needed.
    // Absent/empty means no custom status, never a status to render.
    statusId: t.statusId || undefined,
    statusName: t.statusName || undefined,
    statusColor: t.statusColor || undefined,
  };
}
