/**
 * Pure onboarding utility helpers for the HR onboarding module.
 * All functions are side-effect free.
 */

export type OnboardingTaskStatus = 'not_started' | 'in_progress' | 'completed' | 'skipped';

export interface OnboardingTask {
  id: string;
  title: string;
  status: OnboardingTaskStatus;
  required: boolean;
  dueDate?: string;
}

export function countCompleted(tasks: OnboardingTask[]): number {
  return tasks.filter((t) => t.status === 'completed').length;
}

export function completionPercent(tasks: OnboardingTask[]): number {
  if (tasks.length === 0) return 0;
  return Math.round((countCompleted(tasks) / tasks.length) * 100);
}

export function allRequiredComplete(tasks: OnboardingTask[]): boolean {
  return tasks.filter((t) => t.required).every((t) => t.status === 'completed');
}

export function pendingRequiredTasks(tasks: OnboardingTask[]): OnboardingTask[] {
  return tasks.filter((t) => t.required && t.status !== 'completed');
}

export function taskStatusLabel(status: OnboardingTaskStatus): string {
  const labels: Record<OnboardingTaskStatus, string> = {
    not_started: 'Not Started',
    in_progress: 'In Progress',
    completed: 'Completed',
    skipped: 'Skipped',
  };
  return labels[status];
}

export function isOverdue(task: OnboardingTask, asOfIso: string): boolean {
  if (!task.dueDate) return false;
  if (task.status === 'completed' || task.status === 'skipped') return false;
  return task.dueDate < asOfIso;
}
