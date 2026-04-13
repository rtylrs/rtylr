import type { Task } from '../stores/flowStore';
import type { ApiTask } from '../services/flowApi';

/**
 * Normalises an ApiTask (from the server) into the richer Task shape used
 * by local store consumers and legacy components.  All required Task fields
 * that are optional in ApiTask get safe defaults so callers don't need casts.
 */
export function apiTaskToTask(t: ApiTask): Task {
  return {
    ...t,
    assigneeId: t.assigneeId ?? '',
    identifier: t.identifier ?? t.id,
    createdAt: t.createdAt ?? new Date().toISOString(),
    updatedAt: t.updatedAt ?? new Date().toISOString(),
  };
}
