import type { Task, Project } from '../stores/flowStore';
import { todayIso } from './dateUtils';

/** Converts a task to a flat CSV row string. */
export function taskToCsvRow(task: Task): string {
  const fields = [
    task.identifier,
    `"${task.title.replace(/"/g, '""')}"`,
    task.status,
    task.priority,
    task.assignee,
    task.dueDate,
    task.projectName,
    String(task.estimate ?? ''),
  ];
  return fields.join(',');
}

/** Returns the CSV header row for tasks. */
export function taskCsvHeader(): string {
  return 'Identifier,Title,Status,Priority,Assignee,Due Date,Project,Estimate (h)';
}

/** Converts a list of tasks to a CSV string. */
export function tasksToCsv(tasks: Task[]): string {
  const rows = [taskCsvHeader(), ...tasks.map(taskToCsvRow)];
  return rows.join('\n');
}

/** Converts a project to a flat CSV row string. */
export function projectToCsvRow(project: Project): string {
  const fields = [
    `"${project.name.replace(/"/g, '""')}"`,
    project.health,
    String(project.progress),
    String(project.taskCount),
    String(project.completedCount),
    project.startDate,
    project.targetDate,
  ];
  return fields.join(',');
}

/** Returns the CSV header row for projects. */
export function projectCsvHeader(): string {
  return 'Name,Health,Progress,Tasks,Completed,Start Date,Target Date';
}

/** Converts a list of projects to a CSV string. */
export function projectsToCsv(projects: Project[]): string {
  const rows = [projectCsvHeader(), ...projects.map(projectToCsvRow)];
  return rows.join('\n');
}

/** Creates a data URI for a CSV string. */
export function csvToDataUri(csv: string): string {
  const encoded = encodeURIComponent(csv);
  return `data:text/csv;charset=utf-8,${encoded}`;
}

/** Generates a filename with the given prefix and current date. */
export function exportFilename(prefix: string, ext = 'csv'): string {
  const date = todayIso();
  return `${prefix}-${date}.${ext}`;
}
