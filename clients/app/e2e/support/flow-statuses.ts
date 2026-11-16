import type { Page, Route } from '../fixtures';

/** A project status as the API sends it. Board columns are statuses. */
export interface WireStatus {
  id: string;
  project_id: string;
  name: string;
  bucket: 'backlog' | 'active' | 'done' | 'cancelled';
  legacy_status: string;
  color: string;
  sort_order: number;
  is_default: boolean;
}

/** The default set every new project starts with. */
export function defaultStatuses(projectId: string): WireStatus[] {
  const rows: [string, WireStatus['bucket'], string, string][] = [
    ['Backlog', 'backlog', 'backlog', '#94a3b8'],
    ['To Do', 'backlog', 'todo', '#9ca3af'],
    ['In Progress', 'active', 'in_progress', '#3b82f6'],
    ['In Review', 'active', 'in_review', '#f59e0b'],
    ['Done', 'done', 'done', '#22c55e'],
    ['Cancelled', 'cancelled', 'cancelled', '#ef4444'],
  ];
  return rows.map(([name, bucket, legacy, color], i) => ({
    id: String(i + 1),
    project_id: projectId,
    name,
    bucket,
    legacy_status: legacy,
    color,
    sort_order: i * 10,
    is_default: legacy === 'todo',
  }));
}

/** Answers GET /flow/status/v2/statuses for any project with `statusesFor`. */
export async function stubStatuses(
  page: Page,
  statusesFor: (projectId: string) => WireStatus[] = defaultStatuses
): Promise<void> {
  await page.route(/localhost:9999\/flow\/status\/v2\/statuses(\?.*)?$/, (route: Route) => {
    const projectId = new URL(route.request().url()).searchParams.get('project_id') ?? '1';
    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ statuses: statusesFor(projectId) }),
    });
  });
}
