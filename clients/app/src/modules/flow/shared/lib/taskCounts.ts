/**
 * Tasks still to do: everything not done. The sidebar shows a project's
 * task_count minus completed_count (done), and the project header uses this,
 * so both places give the same number.
 */
export function countOpenTasks(tasks: readonly { status?: string }[]): number {
  return tasks.filter((task) => task.status !== 'done').length;
}
