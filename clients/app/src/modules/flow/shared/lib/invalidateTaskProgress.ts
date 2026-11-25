import type { QueryClient } from '@tanstack/react-query';

/**
 * Every cached view whose numbers come from task status, archiving or cycle
 * membership: task lists, cycle and epic totals, project progress and the
 * Insights dashboards. A write that can change "% done" anywhere calls this
 * so no screen keeps a stale figure until its staleTime lapses.
 */
const TASK_PROGRESS_KEYS = [
  'tasks',
  'my-tasks',
  'task',
  'epic-rollups',
  'cycles',
  'cycle-capacity',
  'milestones',
  'projects',
  'projects-v2',
  'project-overview',
  'insights-dashboard-project',
  'insights-dashboard-projects',
  'insights-dashboard-work',
] as const;

export function invalidateTaskProgress(qc: QueryClient): void {
  for (const key of TASK_PROGRESS_KEYS) {
    void qc.invalidateQueries({ queryKey: [key] });
  }
}
