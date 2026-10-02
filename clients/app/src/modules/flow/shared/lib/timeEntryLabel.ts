/**
 * The label for a logged time entry's project/task pair: the task's own
 * identifier and title when there is a task, the project name alone when
 * there is only a project. Shared by Add Hours (labelling the entry being
 * built) and My day's own entry list (labelling entries already saved) so
 * the same pair renders the same way in both places.
 */
export function formatWorkEntryTarget(
  task: { identifier?: string; title?: string } | undefined,
  project: { name?: string } | undefined
): string {
  if (task) {
    return task.identifier ? `${task.identifier} ${task.title ?? ''}`.trim() : (task.title ?? '');
  }
  return project?.name ?? '';
}
