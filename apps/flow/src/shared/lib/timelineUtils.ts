import type { Task, Milestone, Project } from '../stores/flowStore';

export interface TimelineItem {
  id: string;
  label: string;
  startDate: string;
  endDate: string;
  type: 'task' | 'milestone' | 'project';
  color?: string;
}

/** Converts a Task to a TimelineItem if it has both start and due dates. */
export function taskToTimelineItem(task: Task): TimelineItem | null {
  if (!task.startDate || !task.dueDate) return null;
  return {
    id: task.id,
    label: task.title,
    startDate: task.startDate,
    endDate: task.dueDate,
    type: 'task',
  };
}

/** Converts a Milestone to a TimelineItem (milestone = single date, shown as point). */
export function milestoneToTimelineItem(milestone: Milestone): TimelineItem {
  return {
    id: milestone.id,
    label: milestone.name,
    startDate: milestone.targetDate,
    endDate: milestone.targetDate,
    type: 'milestone',
    color: milestone.project?.color,
  };
}

/** Converts a Project to a TimelineItem. */
export function projectToTimelineItem(project: Project): TimelineItem | null {
  if (!project.startDate || !project.targetDate) return null;
  return {
    id: project.id,
    label: project.name,
    startDate: project.startDate,
    endDate: project.targetDate,
    type: 'project',
    color: project.color,
  };
}

/** Returns timeline items sorted by start date ascending. */
export function sortTimelineItems(items: TimelineItem[]): TimelineItem[] {
  return [...items].sort((a, b) => a.startDate.localeCompare(b.startDate));
}

/** Returns timeline items that overlap with the given date range. */
export function itemsInRange(
  items: TimelineItem[],
  rangeStart: string,
  rangeEnd: string
): TimelineItem[] {
  return items.filter((item) => item.startDate <= rangeEnd && item.endDate >= rangeStart);
}

/** Returns the earliest start date from a list of timeline items. */
export function earliestStart(items: TimelineItem[]): string | null {
  if (items.length === 0) return null;
  return items.reduce(
    (min, item) => (item.startDate < min ? item.startDate : min),
    items[0].startDate
  );
}

/** Returns the latest end date from a list of timeline items. */
export function latestEnd(items: TimelineItem[]): string | null {
  if (items.length === 0) return null;
  return items.reduce((max, item) => (item.endDate > max ? item.endDate : max), items[0].endDate);
}
