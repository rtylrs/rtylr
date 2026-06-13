import { unwrap } from '@voxire/api';
import { useMutation, useQuery } from '@tanstack/react-query';
import { safe } from '../../lib/api';

export interface POSLocation {
  id: number;
  name: string;
  address: string;
}

interface RawLocation {
  id?: number;
  location_id?: number;
  name?: string;
  address?: string;
}

/** Lists the business's POS locations (owner/admin see all). */
export const usePOSLocations = (enabled: boolean) =>
  useQuery<POSLocation[]>({
    queryKey: ['pos-locations'],
    queryFn: async (): Promise<POSLocation[]> => {
      const res = await unwrap(safe.get('/pos/locations'));
      const raw = res as { locations?: RawLocation[] };
      const list = Array.isArray(raw?.locations) ? raw.locations : [];
      return list
        .map((l) => ({
          id: Number(l.id ?? l.location_id ?? 0),
          name: String(l.name ?? ''),
          address: String(l.address ?? ''),
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
export const useAssignPOSLocation = () =>
  useMutation({
    mutationFn: (input: AssignLocationInput) => unwrap(safe.post('/pos/access/assign', input)),
  });

export const useRemovePOSLocation = () =>
  useMutation({
    mutationFn: (id: number) => unwrap(safe.post('/pos/access/remove', { id })),
  });

export interface MemberLocation {
  id: number; // pos_member_locations row id (for removal)
  locationId: number;
  canManagePosUsers: boolean;
  canManageTerminals: boolean;
  canViewReports: boolean;
}

interface RawMemberLocation {
  id?: number;
  location_id?: number;
  locationId?: number;
  can_manage_pos_users?: boolean;
  canManagePosUsers?: boolean;
}

/** A specific team member's POS location assignments (owner/admin only). */
export const useMemberLocations = (userId: string | number | undefined, enabled: boolean) =>
  useQuery<MemberLocation[]>({
    queryKey: ['pos-member-locations', String(userId ?? '')],
    queryFn: async (): Promise<MemberLocation[]> => {
      const res = await unwrap(
        safe.get('/pos/access/member-locations', { params: { user_id: Number(userId) } })
      );
      const raw = res as { locations?: RawMemberLocation[] };
      const list = Array.isArray(raw?.locations) ? raw.locations : [];
      return list
        .map((l) => ({
          id: Number(l.id ?? 0),
          locationId: Number(l.locationId ?? l.location_id ?? 0),
          canManagePosUsers: Boolean(l.canManagePosUsers ?? l.can_manage_pos_users),
          canManageTerminals: false,
          canViewReports: false,
        }))
        .filter((l) => l.locationId > 0);
    },
    enabled: enabled && !!userId,
    staleTime: 30_000,
  });

// store_manager manages its assigned locations; everyone else gets plain access.
// The backend re-derives capabilities, so this only seeds the management flags.
export const locationFlagsForRole = (
  role: string
): { canManagePosUsers: boolean; canManageTerminals: boolean; canViewReports: boolean } =>
  role === 'store_manager'
    ? { canManagePosUsers: true, canManageTerminals: true, canViewReports: true }
    : { canManagePosUsers: false, canManageTerminals: false, canViewReports: false };
