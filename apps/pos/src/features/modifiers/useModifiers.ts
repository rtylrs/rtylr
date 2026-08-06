import { createOneKitClient } from '@voxire/api';
import { ModifierServiceClient } from '@voxire/voxie/pos/modifier/v1/client';
import { useQuery } from '@tanstack/react-query';
import type { ModifierGroup } from '../../shared/types/pos';

const modifierClient = createOneKitClient(ModifierServiceClient, 'pos');

function toModifierGroups(
  modifiers: Awaited<ReturnType<typeof modifierClient.listModifiers>>['modifiers']
): ModifierGroup[] {
  const groups = new Map<string, ModifierGroup>();

  for (const modifier of modifiers ?? []) {
    const groupName = modifier.groupName || 'Modifiers';
    const groupId = groupName.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '');
    const group =
      groups.get(groupName) ??
      ({
        id: groupId || groupName,
        name: groupName,
        min: modifier.isRequired ? 1 : 0,
        max: 0,
        modifiers: [],
      } satisfies ModifierGroup);

    group.min = Math.max(group.min, modifier.isRequired ? 1 : 0);
    group.modifiers.push({
      id: modifier.id ?? '',
      name: modifier.name ?? '',
      priceDelta: modifier.priceDelta ?? 0,
      group: groupName,
      required: modifier.isRequired,
      defaultSelected: false,
    });
    group.max = group.modifiers.length;
    groups.set(groupName, group);
  }

  return Array.from(groups.values());
}

export function useModifiers() {
  const query = useQuery({
    queryKey: ['pos-modifiers'],
    queryFn: async (): Promise<ModifierGroup[]> => {
      const res = await modifierClient.listModifiers({});
      return toModifierGroups(res.modifiers);
    },
    staleTime: 5 * 60 * 1000,
  });

  return {
    groups: query.data ?? [],
    loading: query.isLoading,
    error: query.error
      ? query.error instanceof Error
        ? query.error.message
        : String(query.error)
      : null,
    reload: query.refetch,
  };
}
