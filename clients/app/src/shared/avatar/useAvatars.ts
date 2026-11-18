import { useMemo } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { userClient } from '../lib/clients';
import { useUserStore } from '../stores/userStore';
import type { ChosenAvatar } from './avatarModel';

const KINDS = new Set<ChosenAvatar['kind']>(['photo', 'illustrated', 'initials', 'emoji']);

/** The wire shape (`UserAvatar`) narrowed to what the portal can draw. */
export function toChosenAvatar(
  wire:
    | { kind?: string; url?: string; style?: string; seed?: string; color?: number; emoji?: string }
    | null
    | undefined
): ChosenAvatar | undefined {
  if (!wire?.kind || !KINDS.has(wire.kind as ChosenAvatar['kind'])) return undefined;
  return {
    kind: wire.kind as ChosenAvatar['kind'],
    url: wire.url,
    style: wire.style,
    seed: wire.seed,
    color: wire.color,
    emoji: wire.emoji,
  };
}

export const businessAvatarsKey = (businessId: number | undefined) =>
  ['business-avatars', businessId] as const;

/**
 * Every colleague's chosen avatar, by user id, with the signed-in user's own
 * choice taken from their profile so a change shows at once.
 */
export function useAvatarDirectory(): ReadonlyMap<string, ChosenAvatar> {
  const user = useUserStore((state) => state.user);
  const businessId = user?.businessId;
  const query = useQuery({
    queryKey: businessAvatarsKey(businessId),
    queryFn: async () => {
      const res = await userClient.listBusinessAvatars({});
      return res.avatars ?? [];
    },
    enabled: !!businessId,
    staleTime: 5 * 60 * 1000,
  });

  return useMemo(() => {
    const map = new Map<string, ChosenAvatar>();
    for (const entry of query.data ?? []) {
      const avatar = toChosenAvatar(entry.avatar);
      if (avatar) map.set(String(entry.userId), avatar);
    }
    if (user?.id !== undefined) {
      const own = toChosenAvatar(user.avatar);
      if (own) map.set(String(user.id), own);
      else map.delete(String(user.id));
    }
    return map;
  }, [query.data, user?.id, user?.avatar]);
}

/** Saves the signed-in user's avatar; `null` goes back to the default. */
export function useUpdateAvatar() {
  const queryClient = useQueryClient();
  const setUser = useUserStore((state) => state.setUser);
  return useMutation({
    mutationFn: async (avatar: ChosenAvatar | null) => {
      const res = await userClient.updateAvatar(avatar ? { avatar } : {});
      return toChosenAvatar(res.user?.avatar);
    },
    onSuccess: (saved) => {
      const current = useUserStore.getState().user;
      if (current) setUser({ ...current, avatar: saved });
      void queryClient.invalidateQueries({ queryKey: ['business-avatars'] });
    },
  });
}
