import type { ProjectSection, Task } from '../stores/flowStore';

/** Returns sections for a specific project, sorted by order. */
export function sectionsForProject(
  sections: ProjectSection[],
  projectId: string
): ProjectSection[] {
  return [...sections].filter((s) => s.projectId === projectId).sort((a, b) => a.order - b.order);
}

/** Find a section by its ID. */
export function findSectionById(
  sections: ProjectSection[],
  id: string
): ProjectSection | undefined {
  return sections.find((s) => s.id === id);
}

/** Returns tasks that belong to a specific section. */
export function tasksInSection(tasks: Task[], sectionId: string): Task[] {
  return tasks.filter((t) => t.sectionId === sectionId);
}

/** Returns tasks that have no section (not assigned to any section). */
export function unsectionedTasks(tasks: Task[], projectId: string): Task[] {
  return tasks.filter((t) => t.projectId === projectId && !t.sectionId);
}

/** Returns the next order number for a new section in a project. */
export function nextSectionOrder(sections: ProjectSection[], projectId: string): number {
  const projectSections = sectionsForProject(sections, projectId);
  if (projectSections.length === 0) return 0;
  return Math.max(...projectSections.map((s) => s.order)) + 1;
}

/** Returns true if a section name already exists in the project (case-insensitive). */
export function sectionNameExists(
  sections: ProjectSection[],
  projectId: string,
  name: string
): boolean {
  const lower = name.trim().toLowerCase();
  return sectionsForProject(sections, projectId).some((s) => s.name.toLowerCase() === lower);
}
