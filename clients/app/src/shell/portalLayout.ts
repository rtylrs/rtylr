import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { createOneKitClient } from '@voxire/api';
import { PortalLayoutServiceClient } from '@voxire/voxie/hub/portal/v1/client';
import type {
  PortalArrangement,
  PortalLayout,
  SavePortalLayoutRequest,
} from '@voxire/voxie/hub/portal/v1/types';

export type { PortalArrangement, PortalLayout };

export const portalLayoutClient = createOneKitClient(PortalLayoutServiceClient, 'dash');
export const portalLayoutKey = ['portal-layout'] as const;

// The last layout is kept per person so the app bar doesn't jump when the
// page loads; the server copy replaces it as soon as it arrives.
const cacheKey = (userId: string | number) => `portal-layout:${userId}`;

function readCachedLayout(userId: string | number | undefined): PortalLayout | undefined {
  if (userId === undefined) return undefined;
  try {
    const raw = localStorage.getItem(cacheKey(userId));
    return raw ? (JSON.parse(raw) as PortalLayout) : undefined;
  } catch {
    return undefined;
  }
}

function cacheLayout(userId: string | number | undefined, layout: PortalLayout | undefined) {
  if (userId === undefined || !layout) return;
  try {
    localStorage.setItem(cacheKey(userId), JSON.stringify(layout));
  } catch {
    /* The layout still applies when storage is blocked. */
  }
}

/** The signed-in person's arrangement of the Portal navigation. Undefined
 *  until known, which renders the default. */
export function usePortalLayout(userId: string | number | undefined): PortalLayout | undefined {
  const { data } = useQuery({
    queryKey: [...portalLayoutKey, userId],
    queryFn: async () => {
      const layout = (await portalLayoutClient.getPortalLayout({})).layout ?? {};
      cacheLayout(userId, layout);
      return layout;
    },
    initialData: () => readCachedLayout(userId),
    initialDataUpdatedAt: 0,
    staleTime: 60_000,
    enabled: userId !== undefined,
  });
  return data;
}

export function useSavePortalLayout(userId: string | number | undefined) {
  const queryClient = useQueryClient();
  const key = [...portalLayoutKey, userId];
  return useMutation({
    mutationFn: async (request: SavePortalLayoutRequest) =>
      (await portalLayoutClient.savePortalLayout(request)).layout,
    onMutate: async (request) => {
      await queryClient.cancelQueries({ queryKey: key });
      const previous = queryClient.getQueryData<PortalLayout>(key);
      queryClient.setQueryData<PortalLayout>(key, {
        ...previous,
        ...(request.rail ? { rail: { ...request.rail, customized: true } } : {}),
        ...(request.workspace ? { workspace: { ...request.workspace, customized: true } } : {}),
      });
      return { previous };
    },
    onError: (_error, _request, context) => queryClient.setQueryData(key, context?.previous),
    onSuccess: (layout) => {
      cacheLayout(userId, layout);
      queryClient.setQueryData(key, layout);
    },
  });
}

export function useResetPortalLayout(userId: string | number | undefined) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (area: 'rail' | 'workspace' | '') =>
      (await portalLayoutClient.resetPortalLayout({ area })).layout,
    onSuccess: (layout) => {
      cacheLayout(userId, layout);
      queryClient.setQueryData([...portalLayoutKey, userId], layout);
    },
  });
}

/**
 * Orders navigation items by a saved arrangement. Listed items come first in
 * their saved order and group; items it doesn't list (one this Portal knows
 * but the server doesn't yet) keep their default group, after the others.
 * Hidden items are left out.
 */
export function arrangeItems<T extends { id: string }>(
  items: T[],
  arrangement: PortalArrangement | undefined,
  defaultBottom: readonly string[] = []
): { top: T[]; bottom: T[] } {
  const byId = new Map(items.map((item) => [item.id, item]));
  const hidden = new Set(arrangement?.hidden ?? []);
  const placed = new Set<string>();
  const take = (ids: readonly string[] | undefined) =>
    (ids ?? []).flatMap((id) => {
      const item = byId.get(id);
      if (!item || placed.has(id)) return [];
      placed.add(id);
      return [item];
    });
  const top = take(arrangement?.top);
  const bottom = take(arrangement?.bottom);
  for (const item of items) {
    if (placed.has(item.id)) continue;
    (defaultBottom.includes(item.id) ? bottom : top).push(item);
  }
  const shown = (item: T) => !hidden.has(item.id);
  return { top: top.filter(shown), bottom: bottom.filter(shown) };
}
