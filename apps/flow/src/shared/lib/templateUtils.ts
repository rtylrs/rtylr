import type { TaskTemplate, Task } from '../stores/flowStore';

/** Find a template by its ID. */
export function findTemplateById(templates: TaskTemplate[], id: string): TaskTemplate | undefined {
  return templates.find((t) => t.id === id);
}

/** Sort templates alphabetically by name. */
export function sortTemplatesAlpha(templates: TaskTemplate[]): TaskTemplate[] {
  return [...templates].sort((a, b) => a.name.localeCompare(b.name));
}

/** Returns templates filtered by priority. */
export function templatesByPriority(
  templates: TaskTemplate[],
  priority: Task['priority']
): TaskTemplate[] {
  return templates.filter((t) => t.priority === priority);
}

/** Returns a partial Task object derived from a template (excludes IDs). */
export function taskFromTemplate(
  template: TaskTemplate,
  projectId: string,
  workspaceId: string
): Omit<
  Task,
  | 'id'
  | 'assignee'
  | 'assigneeInitials'
  | 'assigneeId'
  | 'createdAt'
  | 'updatedAt'
  | 'identifier'
  | 'projectName'
> {
  return {
    title: template.title,
    description: template.description,
    priority: template.priority,
    status: template.status,
    labelIds: template.labels ?? [],
    subtaskCount: template.subtasks?.length ?? 0,
    dueDate: '',
    projectId,
    workspaceId,
    estimate: template.estimate,
  };
}

/** Returns true when a template name already exists (case-insensitive). */
export function templateNameExists(templates: TaskTemplate[], name: string): boolean {
  const lower = name.trim().toLowerCase();
  return templates.some((t) => t.name.toLowerCase() === lower);
}

/** Count templates grouped by priority. */
export function countTemplatesByPriority(
  templates: TaskTemplate[]
): Partial<Record<Task['priority'], number>> {
  const result: Partial<Record<Task['priority'], number>> = {};
  for (const t of templates) {
    result[t.priority] = (result[t.priority] ?? 0) + 1;
  }
  return result;
}
