import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { AccessServiceClient } from '@voxire/voxie/pos/access/v1/client';
import { toast } from 'sonner';
import i18next from 'i18next';

const accessClient = createOneKitClient(AccessServiceClient, 'dash');

export interface POSLocation {
  id: number;
  name: string;
  address: string;
}

/** Lists POS locations the current user may manage or work at. */
export const usePOSLocations = (enabled: boolean) =>
  useQuery<POSLocation[]>({
    queryKey: ['pos-locations'],
    queryFn: async (): Promise<POSLocation[]> => {
      const res = await accessClient.getMyLocations({});
      return (res.locations ?? [])
        .map((l) => ({
          id: Number(l.locationId),
          name: l.name ?? '',
          address: l.address ?? '',
        }))
        .filter((l) => l.id > 0);
    },
    enabled,
    staleTime: 60_000,
  });

export interface AssignLocationInput {
  userId: number;
  locationId: number;
  canManagePosUsers: boolean;
  canManageTerminals: boolean;
  canViewReports: boolean;
}

/**
 * Assigns a team member to a POS location (writes pos_member_locations).
 * Idempotent on the backend, so re-assigning updates the management flags.
 */
export const useAssignPOSLocation = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: AssignLocationInput) =>
      accessClient.assignLocation({
        userId: String(input.userId),
        locationId: String(input.locationId),
        canManagePosUsers: input.canManagePosUsers,
        canManageTerminals: input.canManageTerminals,
        canViewReports: input.canViewReports,
      }),
    onError: (error) =>
      toast.error(
        getOneKitErrorMessage(
          error,
          i18next.t('toasts.failedAssignLocation', 'Failed to assign location')
        )
      ),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['pos-member-locations'] });
    },
  });
};

export const useRemovePOSLocation = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => accessClient.removeLocation({ id: String(id) }),
    onError: (error) =>
      toast.error(
        getOneKitErrorMessage(
          error,
          i18next.t('toasts.failedRemoveLocation', 'Failed to remove location')
        )
      ),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['pos-member-locations'] });
    },
  });
};

export interface MemberLocation {
  id: number; // pos_member_locations row id (for removal)
  locationId: number;
  canManagePosUsers: boolean;
  canManageTerminals: boolean;
  canViewReports: boolean;
}

/** A specific team member's POS location assignments (owner/admin only). */
export const useMemberLocations = (userId: string | number | undefined, enabled: boolean) =>
  useQuery<MemberLocation[]>({
    queryKey: ['pos-member-locations', String(userId ?? '')],
    queryFn: async (): Promise<MemberLocation[]> => {
      const res = await accessClient.getMemberLocations({ userId: String(userId) });
      return (res.locations ?? [])
        .map((l) => ({
          id: Number(l.id),
          locationId: Number(l.locationId),
          canManagePosUsers: l.canManagePosUsers ?? false,
          canManageTerminals: l.canManageTerminals ?? false,
          canViewReports: l.canViewReports ?? false,
        }))
        .filter((l) => l.locationId > 0);
    },
    enabled: enabled && !!userId,
    staleTime: 30_000,
  });
