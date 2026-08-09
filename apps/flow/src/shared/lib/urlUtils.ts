/** Returns the path for a task detail page. */
export function taskPath(workspaceId: string, taskId: string): string {
  return `/w/${workspaceId}/tasks/${taskId}`;
}

/** Returns the path for a project detail page. */
export function projectPath(workspaceId: string, projectId: string): string {
  return `/w/${workspaceId}/projects/${projectId}`;
}

/** Returns the path for a project's board view. */
export function projectBoardPath(workspaceId: string, projectId: string): string {
  return `/w/${workspaceId}/projects/${projectId}/board`;
}

/** Returns the path for a cycle detail page. */
export function cyclePath(workspaceId: string, cycleId: string): string {
  return `/w/${workspaceId}/cycles/${cycleId}`;
}

/** Returns the path for a milestone detail page. */
export function milestonePath(workspaceId: string, milestoneId: string): string {
  return `/w/${workspaceId}/milestones/${milestoneId}`;
}

/** Returns the path for a workspace inbox. */
export function inboxPath(workspaceId: string): string {
  return `/w/${workspaceId}/inbox`;
}

/** Returns the path for my tasks in a workspace. */
export function myTasksPath(workspaceId: string): string {
  return `/w/${workspaceId}/my-tasks`;
}

/** Returns the path for workspace settings. */
export function workspaceSettingsPath(workspaceId: string): string {
  return `/w/${workspaceId}/settings`;
}

/** Appends query params to a URL path string. */
export function withQueryParams(
  path: string,
  params: Record<string, string | number | boolean | undefined>
): string {
  const query = Object.entries(params)
    .filter(([, v]) => v !== undefined)
    .map(([k, v]) => `${encodeURIComponent(k)}=${encodeURIComponent(String(v))}`)
    .join('&');
  return query ? `${path}?${query}` : path;
}

/** Parses a query string into a record of key-value pairs. */
export function parseQueryString(qs: string): Record<string, string> {
  const params = new URLSearchParams(qs.startsWith('?') ? qs.slice(1) : qs);
  const result: Record<string, string> = {};
  params.forEach((value, key) => {
    result[key] = value;
  });
  return result;
}
