import type { Project } from '../stores/flowStore';
import type { ApiProject } from '../services/flowApi';
import { ACCENT_DEFAULT } from './colorUtils';

/**
 * The wire contract sends numeric IDs (`id: number`), but the Zustand
 * Flow store expects string IDs (URL params, active-id comparisons).
 * These mappers stringify on the way in so `projects.find(p => p.id ===
 * projectId)` works against a `useParams()` string.
 *
 * Extra server-provided display fields (taskCount, completedCount, etc.)
 * flow through unchanged. Fields Zustand needs but the server doesn't
 * yet populate (memberAvatars, health, progress) get safe defaults.
 */
export function apiProjectToProject(p: ApiProject): Project {
  const taskCount = p.taskCount ?? 0;
  const completedCount = p.completedCount ?? 0;
  return {
    id: String(p.id),
    name: p.name,
    color: p.color || ACCENT_DEFAULT,
    description: p.description,
    status: p.status,
    taskCount,
    completedCount,
    memberAvatars: p.memberAvatars ?? [],
    health: 'on_track',
    progress: taskCount ? Math.round((completedCount / taskCount) * 100) : 0,
    startDate: p.startDate ?? '',
    targetDate: p.dueDate ?? '',
    lead: '',
    icon: p.icon,
    visibleViaReport: p.visibleViaReport ?? false,
  };
}
