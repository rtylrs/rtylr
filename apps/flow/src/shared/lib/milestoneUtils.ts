import type { Milestone } from '../stores/flowStore';

/** Returns true when a milestone is overdue (target date passed and not completed). */
export function isMilestoneOverdue(milestone: Milestone, now = new Date()): boolean {
  if (milestone.status === 'completed') return false;
  const target = new Date(milestone.targetDate);
  target.setHours(23, 59, 59, 999);
  return target < now;
}

/** Returns the number of days until the milestone target date. Negative means overdue. */
export function daysUntilMilestone(milestone: Milestone, now = new Date()): number {
  const target = new Date(milestone.targetDate);
  target.setHours(0, 0, 0, 0);
  const today = new Date(now);
  today.setHours(0, 0, 0, 0);
  return Math.round((target.getTime() - today.getTime()) / 86400000);
}

/** Sort milestones by target date (soonest first). */
export function sortMilestonesByDate(milestones: Milestone[]): Milestone[] {
  return [...milestones].sort(
    (a, b) => new Date(a.targetDate).getTime() - new Date(b.targetDate).getTime()
  );
}

/** Filter milestones by status. */
export function filterMilestonesByStatus(
  milestones: Milestone[],
  status: Milestone['status']
): Milestone[] {
  return milestones.filter((m) => m.status === status);
}

/** Group milestones by their status. */
export function groupMilestonesByStatus(
  milestones: Milestone[]
): Map<Milestone['status'], Milestone[]> {
  const map = new Map<Milestone['status'], Milestone[]>();
  for (const m of milestones) {
    const group = map.get(m.status) ?? [];
    group.push(m);
    map.set(m.status, group);
  }
  return map;
}

/** Progress percentage (0-100) using computedProgress when available. */
export function milestoneProgress(milestone: Milestone): number {
  const prog = milestone.computedProgress ?? milestone.progress ?? 0;
  return Math.min(100, Math.max(0, prog));
}
