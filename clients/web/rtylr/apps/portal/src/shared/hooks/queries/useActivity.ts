import { createOneKitClient } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { ActivityServiceClient } from '@voxire/voxie/hub/activity/v1/client';
import type { ActivityEntry as ApiActivityEntry } from '@voxire/voxie/hub/activity/v1/types';
import { activityEntrySchema, type ActivityEntry } from '../../schemas/activity';

export type { ActivityEntry } from '../../schemas/activity';

const activityClient = createOneKitClient(ActivityServiceClient, 'dash');

// hub.activity.v1.ListActivity has no app filter - the contract carries
// limit + offset for paging through the feed client- or server-side.

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

function mapEntry(raw: ApiActivityEntry): ActivityEntry {
  const entityType = raw.entityType ?? '';
  const action = raw.action ?? '';
  const target =
    (raw.description && raw.description.trim()) ||
    `${humanize(entityType)}${raw.entityId && raw.entityId !== '0' ? ` #${raw.entityId}` : ''}`.trim() ||
    '—';

  return activityEntrySchema.parse({
    id: raw.id,
    actor:
      (raw.userName && raw.userName.trim()) ||
      (raw.userId && raw.userId !== '0' ? `User #${raw.userId}` : 'System'),
    actorId: raw.userId && raw.userId !== '0' ? raw.userId : undefined,
    action: humanize(action) || action || 'activity',
    target,
    app: raw.app || 'hub',
    createdAt: raw.createdAt || new Date().toISOString(),
    severity: deriveSeverity(action),
  });
}

// useActivity returns the most recent `limit` entries as a flat array — used by
// the dashboard activity widget, the activity feed, and member detail.
export const useActivity = (limit = 100) => {
  return useQuery({
    queryKey: ['activity', limit],
    queryFn: async (): Promise<ActivityEntry[]> => {
      const res = await activityClient.listActivity({ limit });
      return (res.activities ?? []).map(mapEntry);
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

export const useAuditLog = (params: { page: number; pageSize: number }) => {
  const { page, pageSize } = params;
  return useQuery<AuditPage>({
    queryKey: ['audit', { page, pageSize }],
    queryFn: async (): Promise<AuditPage> => {
      const res = await activityClient.listActivity({
        offset: (page - 1) * pageSize,
        limit: pageSize,
      });
      const entries = (res.activities ?? []).map(mapEntry);
      const total = res.total ?? entries.length;
      return {
        entries,
        total,
        page,
        pageSize,
      };
    },
    staleTime: 20_000,
    refetchOnWindowFocus: true,
  });
};
