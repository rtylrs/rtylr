import type { Member, Task } from '../stores/flowStore';

/** Capacity utilization as a percentage (0-100+). Values over 100 indicate overload. */
export function capacityUtilizationPercent(assignedHours: number, capacityHours: number): number {
  if (!capacityHours || capacityHours === 0) return 0;
  return Math.round((assignedHours / capacityHours) * 100);
}

/** Returns 'over' | 'near' | 'ok' based on load percentage. */
export function loadTier(utilizationPercent: number): 'over' | 'near' | 'ok' {
  if (utilizationPercent > 100) return 'over';
  if (utilizationPercent >= 80) return 'near';
  return 'ok';
}

/** CSS color for a load tier. */
export function loadTierColor(tier: 'over' | 'near' | 'ok'): string {
  const colors: Record<'over' | 'near' | 'ok', string> = {
    over: 'var(--flow-danger-fg)',
    near: 'var(--flow-warn-fg)',
    ok: 'var(--flow-status-approved)',
  };
  return colors[tier];
}

/** Total estimated hours of open tasks for a given assignee. */
export function totalEstimatedHoursForMember(tasks: Task[], memberId: string): number {
  return tasks
    .filter((t) => t.assigneeId === memberId && t.status !== 'done' && t.status !== 'cancelled')
    .reduce((sum, t) => sum + (t.estimate ?? 0), 0);
}

/** Count of open tasks for a given assignee. */
export function openTaskCountForMember(tasks: Task[], memberId: string): number {
  return tasks.filter(
    (t) => t.assigneeId === memberId && t.status !== 'done' && t.status !== 'cancelled'
  ).length;
}

/** Sort members by their load (highest load first). */
export function sortMembersByLoad(
  members: Array<Member & { assignedHours: number }>
): Array<Member & { assignedHours: number }> {
  return [...members].sort((a, b) => {
    // Unknown capacity sorts as zero load rather than infinite: a member whose
    // schedule is unset is not evidence of being overloaded (#868).
    const loadA = a.capacity && a.capacity > 0 ? a.assignedHours / a.capacity : 0;
    const loadB = b.capacity && b.capacity > 0 ? b.assignedHours / b.capacity : 0;
    return loadB - loadA;
  });
}
