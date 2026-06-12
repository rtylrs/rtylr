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
