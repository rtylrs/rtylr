import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { safe } from '../../lib/api';
import { activityEntrySchema, type ActivityEntry } from '../../schemas/activity';

export type { ActivityEntry } from '../../schemas/activity';

// The backend returns rows shaped { app, entityType, entityId, action,
// description, userId, userName, createdAt } (camelCased by @voxire/api). The
// display schema wants { actor, action, target, severity }, so we map here —
// previously the raw rows were parsed directly and the schema threw on every
// row (missing actor/target), which is why the audit log rendered empty.
interface RawActivity {
  id: number | string;
  app?: string;
  entityType?: string;
  entityId?: number;
  action?: string;
  description?: string;
  userId?: number;
  userName?: string;
  createdAt?: string;
}

interface ActivityResponse {
  activities?: RawActivity[];
  entries?: RawActivity[];
  total?: number;
}

// Strip any "<domain>." prefix and turn separators into spaces:
// "product.created" -> "created", "stock_adjusted" -> "stock adjusted".
function humanize(value: string): string {
  return value
    .replace(/^.*\./, '')
    .replace(/[._-]+/g, ' ')
    .trim();
}

// Audit severity is derived from the action verb since the backend does not
// store a severity column. Destructive actions are critical; auth/permission and
// other state changes are warnings; everything else is informational.
function deriveSeverity(action: string): 'info' | 'warning' | 'critical' {
  const a = action.toLowerCase();
  if (/(delete|remove|revoke|disable|destroy|fail|denied|breach|block)/.test(a)) {
    return 'critical';
  }
  if (/(login|logout|permission|role|password|invite|access|export|deactivat|update)/.test(a)) {
    return 'warning';
  }
  return 'info';
}

function mapRawEntry(raw: RawActivity): ActivityEntry {
  const action = raw.action ?? '';
  const entityType = raw.entityType ?? '';
  const target =
    (raw.description && raw.description.trim()) ||
    `${humanize(entityType)}${raw.entityId ? ` #${raw.entityId}` : ''}`.trim() ||
    '—';

  return activityEntrySchema.parse({
    id: raw.id,
    actor: (raw.userName && raw.userName.trim()) || (raw.userId ? `User #${raw.userId}` : 'System'),
    actorId: raw.userId ? String(raw.userId) : undefined,
    action: humanize(action) || action || 'activity',
    target,
    app: raw.app ?? 'hub',
    createdAt: raw.createdAt ?? new Date().toISOString(),
    severity: deriveSeverity(action),
  });
}

function extractRows(res: ActivityResponse): RawActivity[] {
  const raw = res?.activities ?? res?.entries ?? [];
  return Array.isArray(raw) ? raw : [];
}

// useActivity returns the most recent `limit` entries as a flat array — used by
// the dashboard activity widget, the activity feed, and member detail.
export const useActivity = (limit = 100) => {
  return useQuery({
    queryKey: ['activity', limit],
    queryFn: async (): Promise<ActivityEntry[]> => {
      const res = await unwrap(safe.get<ActivityResponse>(`/activity?limit=${limit}`));
      return extractRows(res).map(mapRawEntry);
    },
    staleTime: 20_000,
  });
};

export interface AuditPage {
  entries: ActivityEntry[];
  total: number;
  page: number;
  pageSize: number;
}

// useAuditLog is server-side paginated: it requests one page (limit/offset) and
// reads the total count the backend returns so the UI can render page controls.
export const useAuditLog = (params: { page: number; pageSize: number; app?: string }) => {
  const { page, pageSize, app } = params;
  const offset = Math.max(0, (page - 1) * pageSize);
  return useQuery({
    queryKey: ['audit', page, pageSize, app ?? ''],
    queryFn: async (): Promise<AuditPage> => {
      const search = new URLSearchParams({ limit: String(pageSize), offset: String(offset) });
      if (app) search.set('app', app);
      const res = await unwrap(safe.get<ActivityResponse>(`/activity?${search.toString()}`));
      const rows = extractRows(res);
      const total = typeof res?.total === 'number' ? res.total : rows.length;
      return { entries: rows.map(mapRawEntry), total, page, pageSize };
    },
    staleTime: 20_000,
    placeholderData: (prev) => prev,
  });
};
