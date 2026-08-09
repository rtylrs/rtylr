// Onboarding task helpers for the recruit admin panel.

export type OnboardingTaskStatus = 'pending' | 'in_progress' | 'completed' | 'skipped';

export interface OnboardingTask {
  id: string;
  title: string;
  status: OnboardingTaskStatus;
  dueDate?: string;
  assignedTo?: string;
}

/**
 * onboardingTaskStatusLabel returns a display label for a task status.
 */
export function onboardingTaskStatusLabel(status: OnboardingTaskStatus | string): string {
  const labels: Record<string, string> = {
    pending: 'Pending',
    in_progress: 'In Progress',
    completed: 'Completed',
    skipped: 'Skipped',
  };
  return labels[status] ?? status;
}

/**
 * completedTasks returns tasks with completed status.
 */
export function completedTasks(tasks: OnboardingTask[]): OnboardingTask[] {
  return tasks.filter((t) => t.status === 'completed');
}

/**
 * pendingTasks returns tasks that are pending or in progress.
 */
export function pendingTasks(tasks: OnboardingTask[]): OnboardingTask[] {
  return tasks.filter((t) => t.status === 'pending' || t.status === 'in_progress');
}

/**
 * onboardingProgress returns the percentage of tasks completed.
 */
export function onboardingProgress(tasks: OnboardingTask[]): number {
  const active = tasks.filter((t) => t.status !== 'skipped');
  if (active.length === 0) return 0;
  return Math.round((completedTasks(tasks).length / active.length) * 1000) / 10;
}

/**
 * isOnboardingComplete returns true when all non-skipped tasks are completed.
 */
export function isOnboardingComplete(tasks: OnboardingTask[]): boolean {
  return pendingTasks(tasks).length === 0;
}

/**
 * tasksDueToday returns tasks with a dueDate matching the given date string.
 */
export function tasksDueToday(tasks: OnboardingTask[], today: string): OnboardingTask[] {
  return tasks.filter((t) => t.dueDate === today && t.status !== 'completed');
}
