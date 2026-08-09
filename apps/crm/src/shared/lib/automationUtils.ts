import type { Automation } from '../schemas/automations';

/** Returns automations filtered by status. */
export function filterByStatus(automations: Automation[], status: string): Automation[] {
  return automations.filter((a) => a.status === status);
}

/** Returns the completion rate as a fraction (0-1) for an automation. */
export function completionRate(automation: Automation): number {
  if (automation.enrolled === 0) return 0;
  return automation.completed / automation.enrolled;
}

/** Returns the number of customers still enrolled but not yet completed. */
export function inProgress(automation: Automation): number {
  return Math.max(0, automation.enrolled - automation.completed);
}

/** Returns true if the automation has any enrolled customers. */
export function hasEnrolled(automation: Automation): boolean {
  return automation.enrolled > 0;
}

/** Returns automations sorted by enrolled count, highest first. */
export function sortByEnrollment(automations: Automation[]): Automation[] {
  return [...automations].sort((a, b) => b.enrolled - a.enrolled);
}

/** Returns the total enrolled count across all automations. */
export function totalEnrolled(automations: Automation[]): number {
  return automations.reduce((sum, a) => sum + a.enrolled, 0);
}
