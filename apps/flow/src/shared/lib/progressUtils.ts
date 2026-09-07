import type { Project, Milestone } from '../stores/flowStore';

/** Completion percentage (0-100) for a project based on its task counts. */
export function projectCompletionPercent(project: Project): number {
  if (!project.taskCount || project.taskCount === 0) return 0;
  return Math.min(100, Math.round((project.completedCount / project.taskCount) * 100));
}

/** Completion percentage (0-100) for a milestone. */
export function milestoneCompletionPercent(milestone: Milestone): number {
  if (!milestone.taskCount || milestone.taskCount === 0) return 0;
  return Math.min(100, Math.round((milestone.doneCount / milestone.taskCount) * 100));
}

/** Returns the health label for a project health value. */
export function projectHealthLabel(health: Project['health']): string {
  const labels: Record<Project['health'], string> = {
    on_track: 'On Track',
    at_risk: 'At Risk',
    off_track: 'Off Track',
  };
  return labels[health] ?? health;
}

/** Returns a CSS color for a project health value. */
export function projectHealthColor(health: Project['health']): string {
  const colors: Record<Project['health'], string> = {
    on_track: 'var(--flow-status-approved)',
    at_risk: 'var(--flow-warn-fg)',
    off_track: 'var(--flow-danger-fg)',
  };
  return colors[health] ?? 'var(--flow-neutral-fg)';
}

/** Infer health from completion percent: on_track >= 75, at_risk >= 40, else off_track. */
export function inferProjectHealth(completionPercent: number): Project['health'] {
  if (completionPercent >= 75) return 'on_track';
  if (completionPercent >= 40) return 'at_risk';
  return 'off_track';
}

/** Aggregate total progress across an array of projects (weighted by task count). */
export function aggregateProgress(projects: Project[]): number {
  const totalTasks = projects.reduce((sum, p) => sum + (p.taskCount ?? 0), 0);
  if (totalTasks === 0) return 0;
  const totalCompleted = projects.reduce((sum, p) => sum + (p.completedCount ?? 0), 0);
  return Math.min(100, Math.round((totalCompleted / totalTasks) * 100));
}
